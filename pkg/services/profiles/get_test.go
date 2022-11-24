package profiles_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/services/profiles"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

const getProfilesResp = `{
	"profiles": [
		{
			"name": "podinfo",
			"versions": [
				"6.0.0",
				"6.0.1"
			],
			"layer": "default",
			"repository": {
				"name": "weaveworks-charts",
				"namespace": "flux-system",
				"kind": "HelmRepository",
				"cluster": {
					"name": "dev",
					"namespace": "default"
				}
			}
		}
	]
}`

var _ = Describe("Get Profile(s)", func() {
	var (
		buffer      *gbytes.Buffer
		profilesSvc *profiles.ProfilesSvc
		fakeLogger  logger.Logger
	)

	BeforeEach(func() {
		buffer = gbytes.NewBuffer()
		fakeLogger = logger.From(logr.Discard())
		profilesSvc = profiles.NewService(fakeLogger)
	})

	Context("Get", func() {
		It("prints the available profiles", func() {
			client := NewFakeHTTPClient(getProfilesResp, nil)

			Expect(profilesSvc.Get(context.TODO(), client, buffer, profiles.GetOptions{})).To(Succeed())

			Expect(string(buffer.Contents())).To(Equal(`NAME	AVAILABLE_VERSIONS	LAYER
podinfo	6.0.0,6.0.1	default
`))
		})

		When("the response isn't valid", func() {
			It("errors", func() {
				client := NewFakeHTTPClient("not=json", nil)

				err := profilesSvc.Get(context.TODO(), client, buffer, profiles.GetOptions{})
				Expect(err).To(MatchError(ContainSubstring("failed to unmarshal response")))
			})
		})

		When("making the request fails", func() {
			It("errors", func() {
				client := NewFakeHTTPClient("", fmt.Errorf("nope"))

				err := profilesSvc.Get(context.TODO(), client, buffer, profiles.GetOptions{})
				Expect(err).To(MatchError("unable to retrieve profiles from \"Fake Client\": nope"))
			})
		})

		When("the request returns non-200", func() {
			It("errors", func() {
				client := NewFakeHTTPClient("", &errors.StatusError{ErrStatus: metav1.Status{Code: http.StatusNotFound}})

				err := profilesSvc.Get(context.TODO(), client, buffer, profiles.GetOptions{})
				Expect(err).To(MatchError("unable to retrieve profiles from \"Fake Client\": status code 404"))
			})
		})
	})

	Context("GetProfile", func() {
		var (
			opts profiles.GetOptions
		)

		BeforeEach(func() {
			opts = profiles.GetOptions{
				Name:      "podinfo",
				Version:   "latest",
				Cluster:   "prod",
				Namespace: "test-namespace",
			}
		})

		It("returns an available profile", func() {
			client := NewFakeHTTPClient(getProfilesResp, nil)

			profile, version, err := profilesSvc.GetProfile(context.TODO(), client, opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(profile.Versions)).NotTo(BeZero())
			Expect(version).To(Equal("6.0.1"))
		})

		It("fails to return a list of available profiles from the cluster", func() {
			client := NewFakeHTTPClient("", fmt.Errorf("nope"))

			_, _, err := profilesSvc.GetProfile(context.TODO(), client, opts)
			Expect(err).To(MatchError("unable to retrieve profiles from \"Fake Client\": nope"))
		})

		It("fails if no available profile was found that matches the name for the profile being added", func() {
			badProfileResp := `{
				"profiles": [
				  {
					"name": "foo"
				  }
				]
			  }
			  `

			client := NewFakeHTTPClient(badProfileResp, nil)

			_, _, err := profilesSvc.GetProfile(context.TODO(), client, opts)
			Expect(err).To(MatchError("no available profile 'podinfo' found in prod/test-namespace"))
		})

		It("fails if no available profile was found that matches the name for the profile being added", func() {
			badProfileResp := `{
				"profiles": [
				  {
					"name": "podinfo",
					"availableVersions": [
					]
				  }
				]
			  }
			  `

			client := NewFakeHTTPClient(badProfileResp, nil)
			_, _, err := profilesSvc.GetProfile(context.TODO(), client, opts)
			Expect(err).To(MatchError("no version found for profile 'podinfo' in prod/test-namespace"))
		})

		It("fails if the available profile's HelmRepository name or namespace are empty", func() {
			badProfileResp := `{
				"profiles": [
				  {
					"name": "podinfo",
					"versions": [
					  "6.0.0",
					  "6.0.1"
					],
					"layer": "default",
					"repository": {
						"name": "",
						"namespace": "",
						"kind": "HelmRepository",
						"cluster": {
							"name": "dev",
							"namespace": "default"
						}
					}
				  }
				]
			  }
			  `

			client := NewFakeHTTPClient(badProfileResp, nil)

			_, _, err := profilesSvc.GetProfile(context.TODO(), client, opts)
			Expect(err).To(MatchError("HelmRepository's name or namespace is empty"))
		})
	})
})

type FakeHTTPClient struct {
	data string
	err  error
}

func NewFakeHTTPClient(data string, err error) *FakeHTTPClient {
	return &FakeHTTPClient{data, err}
}

func (c *FakeHTTPClient) Source() string {
	return "Fake Client"
}

func (c *FakeHTTPClient) RetrieveProfiles(req profiles.GetOptions) (profiles.Profiles, error) {
	if c.err != nil {
		return profiles.Profiles{}, c.err
	}

	result := profiles.Profiles{}
	data := []byte(c.data)

	err := json.Unmarshal(data, &result)
	if err != nil {
		return profiles.Profiles{}, fmt.Errorf("failed to unmarshal response")
	}

	return result, nil
}
