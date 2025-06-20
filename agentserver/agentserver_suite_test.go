package agentserver_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAgentserver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Agentserver Suite")
}
