package client_test

import (
	"fmt"
	"github.com/golang/mock/gomock"
	_ "github.com/golang/mock/mockgen/model"
	"github.com/kardolus/maps/client"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"testing"
)

//go:generate mockgen -destination=callermocks_test.go -package=client_test github.com/kardolus/maps/http Caller

const apiKey = "api-key"

var (
	mockCtrl   *gomock.Controller
	mockCaller *MockCaller
	subject    *client.Client
)

const (
	singlePageResponse = `{
		"results": [
			{
				"geometry": {
					"location": {
						"lat": 1.0,
						"lng": 2.0
					}
				}
			}
		],
		"status": "OK"
	}`

	multiPageResponse = `{
		"results": [
			{
				"geometry": {
					"location": {
						"lat": 3.0,
						"lng": 4.0
					}
				}
			}
		],
		"next_page_token": "next-page-token",
		"status": "OK"
	}`
)

func TestUnitClient(t *testing.T) {
	spec.Run(t, "Client Package Unit Tests", testClient, spec.Report(report.Terminal{}))
}

func testClient(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
		mockCtrl = gomock.NewController(t)
		mockCaller = NewMockCaller(mockCtrl)

		subject = client.New(mockCaller, apiKey)
	})

	it.After(func() {
		mockCtrl.Finish()
	})

	when("FetchLocations()", func() {

		it("returns an error when the entity is empty", func() {
			_, err := subject.FetchLocations("")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(client.ErrMissingEntity))
		})

		it("constructs the correct URL for a single entity", func() {
			entity := "New York"
			transformedEntity := "New+York"
			expectedURL := fmt.Sprintf(client.Endpoint, transformedEntity, apiKey)

			mockCaller.EXPECT().Get(expectedURL).Return([]byte(`{}`), nil)

			_, err := subject.FetchLocations(entity)
			Expect(err).NotTo(HaveOccurred())
		})

		it("fetches locations with a single-page result", func() {
			entity := "New York"
			transformedEntity := "New+York"
			expectedURL := fmt.Sprintf(client.Endpoint, transformedEntity, apiKey)

			mockCaller.EXPECT().Get(expectedURL).Return([]byte(singlePageResponse), nil)

			result, err := subject.FetchLocations(entity)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(HaveLen(1))
			Expect(result[0].Results[0].Geometry.Location.Lat).To(Equal(1.0))
		})

		it("fetches locations with a multi-page result", func() {
			entity := "New York"
			transformedEntity := "New+York"
			expectedURL := fmt.Sprintf(client.Endpoint, transformedEntity, apiKey)

			// First page
			mockCaller.EXPECT().Get(expectedURL).Return([]byte(multiPageResponse), nil).Times(1)

			// Second page
			expectedNextPageURL := fmt.Sprintf(client.NextPageEndpoint, "next-page-token", apiKey)
			mockCaller.EXPECT().Get(expectedNextPageURL).Return([]byte(singlePageResponse), nil).Times(1)

			result, err := subject.FetchLocations(entity)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(HaveLen(2))
			Expect(result[0].Results[0].Geometry.Location.Lat).To(Equal(3.0))
			Expect(result[1].Results[0].Geometry.Location.Lat).To(Equal(1.0))
		})
	})
}
