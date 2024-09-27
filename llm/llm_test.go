package llm_test

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kardolus/maps/llm"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var (
	mockCtrl   *gomock.Controller
	mockClient *MockLLMClient
	mockReader *MockFileReader
	subject    *llm.LLM
)

func TestUnitLLM(t *testing.T) {
	spec.Run(t, "LLM Package Unit Tests", testLLM, spec.Report(report.Terminal{}))
}

func testLLM(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
		mockCtrl = gomock.NewController(t)
		mockClient = NewMockLLMClient(mockCtrl)
		mockReader = NewMockFileReader(mockCtrl)

		subject = llm.New(mockClient, mockReader)
	})

	it.After(func() {
		mockCtrl.Finish()
	})

	it("generates sub-queries for a valid input query", func() {
		query := "Whole Foods in USA"
		mockReader.EXPECT().FileToBytes("query_prompt.txt").Return([]byte("prompt content"), nil)
		mockClient.EXPECT().ProvideContext("prompt content")
		mockClient.EXPECT().Query("input query: "+query).Return("search [1]: Whole Foods in New York\nsearch [2]: Whole Foods in California", 0, nil)

		result, err := subject.GenerateSubQueries(query)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]string{"Whole Foods in New York", "Whole Foods in California"}))
	})

	it("extracts 'contains' and 'matches' from a query", func() {
		query := "Whole Foods in USA"
		mockReader.EXPECT().FileToBytes("filter_prompt.txt").Return([]byte("filter prompt content"), nil)
		mockClient.EXPECT().ProvideContext("filter prompt content")
		mockClient.EXPECT().Query("input query: "+query).Return("contains: whole foods market, whole foods\nmatches: whole foods", 0, nil)

		contains, matches, err := subject.GenerateFilter(query)

		Expect(err).NotTo(HaveOccurred())
		Expect(contains).To(Equal([]string{"whole foods market", "whole foods"}))
		Expect(matches).To(Equal([]string{"whole foods"}))
	})

	it("returns an error when FileToBytes fails in GenerateSubQueries", func() {
		mockReader.EXPECT().FileToBytes("query_prompt.txt").Return(nil, fmt.Errorf("file read error"))

		_, err := subject.GenerateSubQueries("Whole Foods in USA")

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("file read error"))
	})

	it("returns an error when Query fails in GenerateSubQueries", func() {
		mockReader.EXPECT().FileToBytes("query_prompt.txt").Return([]byte("prompt content"), nil)
		mockClient.EXPECT().ProvideContext("prompt content")
		mockClient.EXPECT().Query("input query: Whole Foods in USA").Return("", 0, fmt.Errorf("query error"))

		_, err := subject.GenerateSubQueries("Whole Foods in USA")

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("query error"))
	})
}
