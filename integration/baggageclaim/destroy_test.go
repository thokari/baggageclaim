package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/concourse/baggageclaim"
)

var _ = Describe("Destroying", func() {
	var (
		runner *BaggageClaimRunner
		client baggageclaim.Client
	)

	BeforeEach(func() {
		runner = NewRunner(baggageClaimPath)
		runner.Start()

		client = runner.Client()
	})

	AfterEach(func() {
		runner.Stop()
		runner.Cleanup()
	})

	It("destroys volume", func() {
		createdVolume, err := client.CreateVolume(logger, "some-handle", baggageclaim.VolumeSpec{})
		Expect(err).NotTo(HaveOccurred())

		Expect(runner.CurrentHandles()).To(ConsistOf(createdVolume.Handle()))

		err = createdVolume.Destroy()
		Expect(err).NotTo(HaveOccurred())

		Expect(runner.CurrentHandles()).NotTo(ConsistOf(createdVolume.Handle()))
	})
})
