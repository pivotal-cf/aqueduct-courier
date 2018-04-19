package ops_test

import (
	. "github.com/pivotal-cf/aqueduct-courier/ops"

	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/aqueduct-courier/ops/opsfakes"
	"github.com/pivotal-cf/aqueduct-courier/opsmanager"
)

var _ = Describe("Collector", func() {
	var (
		dataCollector *opsfakes.FakeDataCollector
		writer        *opsfakes.FakeWriter
		collector     CollectExecutor
	)

	BeforeEach(func() {
		dataCollector = new(opsfakes.FakeDataCollector)
		writer = new(opsfakes.FakeWriter)

		collector = NewCollector(dataCollector, writer)
	})

	It("collects data and writes it", func() {
		d1 := opsmanager.NewData(nil, "d1", "best-kind")
		d2 := opsmanager.NewData(nil, "d2", "better-kind")
		dataToWrite := []opsmanager.Data{d1, d2}
		dataCollector.CollectReturns(dataToWrite, nil)

		err := collector.Collect("some/path")
		Expect(err).NotTo(HaveOccurred())
		Expect(writer.MkdirCallCount()).To(Equal(1))
		Expect(writer.MkdirArgsForCall(0)).To(Equal("some/path"))
		Expect(writer.WriteCallCount()).To(Equal(2))
		Expect(writer.WriteArgsForCall(0)).To(Equal(d1))
		Expect(writer.WriteArgsForCall(1)).To(Equal(d2))
	})

	It("returns an error when the collection errors", func() {
		dataCollector.CollectReturns([]opsmanager.Data{}, errors.New("collecting is hard"))

		err := collector.Collect("")
		Expect(err).To(MatchError(ContainSubstring(CollectFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("collecting is hard")))
	})

	It("returns an error when the folder creation fails", func() {
		writer.MkdirReturns("", errors.New("directories are hard"))

		err := collector.Collect("")
		Expect(err).To(MatchError(ContainSubstring(DirCreateFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("directories are hard")))
	})

	It("returns an error when the writing the data fails", func() {
		dataCollector.CollectReturns([]opsmanager.Data{{}}, nil)
		writer.WriteReturns(errors.New("writing datas is hard"))

		err := collector.Collect("")
		Expect(err).To(MatchError(ContainSubstring(DataWriteFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("writing datas is hard")))
	})
})
