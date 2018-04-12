package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAqueductCollector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AqueductCourier Suite")
}
