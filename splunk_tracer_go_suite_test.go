package splunktracing_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSplunkTracerGo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SplunkTracerGo Suite")
}
