package opsmanager_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOpsmanager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Opsmanager Suite")
}
