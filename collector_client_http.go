package splunktracing

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"gzip"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

var (
	acceptHeader      		= http.CanonicalHeaderKey("Accept")
	contentTypeHeader 		= http.CanonicalHeaderKey("Content-Type")
	contentEncodingHeader 	= http.CanonicalHeaderKey("Content-Encoding")
	authHeader        		= http.CanonicalHeaderKey("Authorization")
)

const (
	collectorHTTPMethod = "POST"
	collectorHTTPPath   = "/services/collector"
	contentType    		= "application/json"
	contentEncoding 	= "gzip"
)

type httpCollectorClient struct {
	// auth and runtime information
	reporterID  uint64
	accessToken string // accessToken is the access token used for explicit trace collection requests.
	attributes  map[string]string

	tlsClientConfig *tls.Config
	reportTimeout   time.Duration
	reportingPeriod time.Duration

	// Remote service that will receive reports.
	url    *url.URL
	client *http.Client

	// converters
	converter *hecConverter
}

type transportCloser struct {
	*http.Transport
}

func (closer transportCloser) Close() error {
	closer.CloseIdleConnections()
	return nil
}

type hecReportResponse struct {
	text string
	code int
}

func newHTTPCollectorClient(
	opts Options,
	reporterID uint64,
	attributes map[string]string,
) (*httpCollectorClient, error) {
	url, err := url.Parse(opts.Collector.URL())
	if err != nil {
		fmt.Println("collector config does not produce valid url", err)
		return nil, err
	}
	url.Path = collectorHTTPPath

	tlsClientConfig, err := getTLSConfig(opts.Collector.CustomCACertFile)
	if err != nil {
		fmt.Println("failed to get TLSConfig: ", err)
		return nil, err
	}

	return &httpCollectorClient{
		reporterID:      reporterID,
		accessToken:     opts.AccessToken,
		attributes:      attributes,
		tlsClientConfig: tlsClientConfig,
		reportTimeout:   opts.ReportTimeout,
		reportingPeriod: opts.ReportingPeriod,
		url:             url,
		converter:       newHECConverter(opts),
	}, nil
}

// getTLSConfig returns a *tls.Config according to whether a user has supplied a customCACertFile. If they have,
// we return a TLSConfig that uses the custom CA cert as the lone Root CA. If not, we return nil which http.Transport
// will interpret as the default system defined Root CAs.
func getTLSConfig(customCACertFile string) (*tls.Config, error) {
	if len(customCACertFile) == 0 {
		return nil, nil
	}

	caCerts := x509.NewCertPool()
	cert, err := ioutil.ReadFile(customCACertFile)
	if err != nil {
		return nil, err
	}

	if !caCerts.AppendCertsFromPEM(cert) {
		return nil, fmt.Errorf("credentials: failed to append certificate")
	}

	return &tls.Config{RootCAs: caCerts}, nil
}

func (client *httpCollectorClient) ConnectClient() (Connection, error) {
	// Use a transport independent from http.DefaultTransport to provide sane
	// defaults that make sense in the context of the splunk client. The
	// differences are mostly on setting timeouts based on the report timeout
	// and period.
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   client.reportTimeout / 2,
			DualStack: true,
		}).DialContext,
		// The collector responses are very small, there is no point asking for
		// a compressed payload, explicitly disabling it.
		DisableCompression:     true,
		IdleConnTimeout:        2 * client.reportingPeriod,
		TLSHandshakeTimeout:    client.reportTimeout / 2,
		ResponseHeaderTimeout:  client.reportTimeout,
		ExpectContinueTimeout:  client.reportTimeout,
		MaxResponseHeaderBytes: 64 * 1024, // 64 KB, just a safeguard
		TLSClientConfig:        client.tlsClientConfig,
	}

	client.client = &http.Client{
		Transport: transport,
		Timeout:   client.reportTimeout,
	}

	return transportCloser{transport}, nil
}

func (client *httpCollectorClient) ShouldReconnect() bool {
	// http.Transport will handle connection reuse under the hood
	return false
}

func (client *httpCollectorClient) Report(context context.Context, req reportRequest) (collectorResponse, error) {
	if req.httpRequest == nil {
		return nil, fmt.Errorf("httpRequest cannot be null")
	}

	req.httpRequest.Header.Add(accessTokenHeader, fmt.Sprintf("Splunk %s", client.accessToken))
	httpResponse, err := client.client.Do(req.httpRequest.WithContext(context))
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	response, err := client.toResponse(httpResponse)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (client *httpCollectorClient) Translate(ctx context.Context, buffer *reportBuffer) (reportRequest, error) {

	httpRequest, err := client.toRequest(ctx, buffer)
	if err != nil {
		return reportRequest{}, err
	}
	return reportRequest{
		httpRequest: httpRequest,
	}, nil
}

func (client *httpCollectorClient) toRequest(
	context context.Context,
	buffer *reportBuffer,
) (*http.Request, error) {
	hecRequest := client.converter.toReportRequest(
		client.reporterID,
		client.attributes,
		buffer,
	)

	var outbuf bytes.Buffer
	gz := gzip.NewWriter(outbuf)
    json.NewEncoder(gz).Encode(hecRequest)
    gz.Close()
	requestBody := bytes.NewReader(outbuf)

	request, err := http.NewRequest(collectorHTTPMethod, client.url.String(), requestBody)
	if err != nil {
		return nil, err
	}
	request = request.WithContext(context)
	request.Header.Set(authHeader, "Splunk "+client.accessToken)
	request.Header.Set(contentTypeHeader, contentType)
	request.Header.Set(contentEncodingHeader, contentEncoding)
	request.Header.Set(acceptHeader, contentType)

	return request, nil
}

func (client *httpCollectorClient) toResponse(response *http.Response) (collectorResponse, error) {
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code (%d) is not ok", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	resp := hecReportResponse{}
	if err := json.Unmarshal(body, resp); err != nil {
		return nil, err
	}

	return hecResponse{ReportResponse: resp}, nil
}
