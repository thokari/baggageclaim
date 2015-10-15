package volume_test

import (
	"errors"

	"github.com/concourse/baggageclaim/volume"
	"github.com/concourse/baggageclaim/volume/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Creator", func() {
	var (
		strategyProvider  *fakes.FakeStrategyProvider
		volumeRepo        *fakes.FakeRepository
		creator           volume.Creator
		defaultRootFsPath string
	)

	JustBeforeEach(func() {
		strategyProvider = new(fakes.FakeStrategyProvider)
		volumeRepo = new(fakes.FakeRepository)
		creator = volume.NewCreator(strategyProvider, volumeRepo, defaultRootFsPath)
	})

	Context("when the path is empty", func() {
		BeforeEach(func() {
			defaultRootFsPath = "/this/is/default/root/fs"
		})

		It("should use the default rootfs", func() {
			_, err := creator.Create("")
			Expect(err).NotTo(HaveOccurred())

			Expect(strategyProvider.ProvideStrategyCallCount()).To(Equal(1))
			Expect(strategyProvider.ProvideStrategyArgsForCall(0)).To(Equal(defaultRootFsPath))
		})
	})

	It("returns the volume path from volume repository", func() {
		volumeRepo.CreateVolumeReturns(volume.Volume{Path: "/path/to/banana/rootfs"}, nil)

		path, err := creator.Create("/my/local/root/fs")
		Expect(err).NotTo(HaveOccurred())

		Expect(strategyProvider.ProvideStrategyCallCount()).To(Equal(1))
		Expect(strategyProvider.ProvideStrategyArgsForCall(0)).To(Equal("/my/local/root/fs"))

		Expect(volumeRepo.CreateVolumeCallCount()).To(Equal(1))
		Expect(path).To(Equal("/path/to/banana/rootfs"))
	})

	Context("when the VolumeRepository fails", func() {
		JustBeforeEach(func() {
			volumeRepo.CreateVolumeStub = func(_ volume.Strategy,
				_ volume.Properties, _ uint) (volume.Volume, error) {
				return volume.Volume{}, errors.New("Explode!")
			}
		})

		It("returns a sensible error", func() {
			_, err := creator.Create("/coool/root/fs")
			Expect(err).To(MatchError(("Explode!")))
		})
	})

	It("correctly delegates to the strategyProvider", func() {
		strategy := volume.EmptyStrategy{}

		strategyProvider.ProvideStrategyReturns(strategy, nil)

		_, err := creator.Create("/orig/rootfs")
		Expect(err).NotTo(HaveOccurred())

		Expect(strategyProvider.ProvideStrategyCallCount()).To(Equal(1))
		Expect(strategyProvider.ProvideStrategyArgsForCall(0)).To(Equal("/orig/rootfs"))

		Expect(volumeRepo.CreateVolumeCallCount()).To(Equal(1))
		actualStrategy, _, _ := volumeRepo.CreateVolumeArgsForCall(0)
		Expect(actualStrategy).To(Equal(strategy))
	})

	Context("when the StrategyProvider fails", func() {
		JustBeforeEach(func() {
			strategyProvider.ProvideStrategyReturns(nil, errors.New("So many wombles!"))
		})

		It("returns a sensible error", func() {
			_, err := creator.Create("/my/path")
			Expect(err).To(MatchError("So many wombles!"))
		})

		It("does not create a volume", func() {
			creator.Create("/your/path")
			Expect(volumeRepo.CreateVolumeCallCount()).To(Equal(0))
		})
	})
})
