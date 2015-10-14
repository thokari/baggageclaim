package volume_test

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os/exec"

	"github.com/cloudfoundry/gunk/command_runner/fake_command_runner"
	. "github.com/cloudfoundry/gunk/command_runner/fake_command_runner/matchers"
	"github.com/concourse/baggageclaim/volume"
	"github.com/concourse/baggageclaim/volume/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GardenStrategyProvider", func() {
	var (
		filesystem       *fakes.FakeFilesystem
		initVolume       *fakes.FakeFilesystemInitVolume
		liveVolume       *fakes.FakeFilesystemLiveVolume
		runner           *fake_command_runner.FakeCommandRunner
		strategyProvider volume.StrategyProvider
		rootFsPath       string
	)

	BeforeEach(func() {
		filesystem = new(fakes.FakeFilesystem)
		initVolume = new(fakes.FakeFilesystemInitVolume)
		initVolume.DataPathReturns("/path/to/volume")
		filesystem.NewVolumeReturns(initVolume, nil)

		liveVolume = new(fakes.FakeFilesystemLiveVolume)
		initVolume.InitializeReturns(liveVolume, nil)

		runner = fake_command_runner.New()

		strategyProvider = volume.NewGardenStrategyProvider(filesystem, runner)

		rootFsPath = "/path/to/banana/rootfs"
	})

	Describe("ProvideStrategy", func() {
		Describe("for a directory rootfs", func() {
			Context("when it is used for the fist time", func() {
				It("creates a volume for the rootfs", func() {
					_, err := strategyProvider.ProvideStrategy(rootFsPath)
					Expect(err).NotTo(HaveOccurred())

					Expect(filesystem.NewVolumeCallCount()).To(Equal(1))
					Expect(filesystem.NewVolumeArgsForCall(0)).To(Equal(
						fmt.Sprintf("%x", sha256.Sum256([]byte(rootFsPath)))))
				})

				Context("when volume creation fails", func() {
					BeforeEach(func() {
						filesystem.NewVolumeReturns(nil, errors.New("Ermagahd! An Eeeeeror!"))
					})

					It("fails with a sensible error message", func() {
						_, err := strategyProvider.ProvideStrategy(rootFsPath)
						Expect(err).To(MatchError("Ermagahd! An Eeeeeror!"))
					})
				})

				It("inititalizes the created volume", func() {
					_, err := strategyProvider.ProvideStrategy(rootFsPath)
					Expect(err).NotTo(HaveOccurred())

					Expect(initVolume.InitializeCallCount()).To(Equal(1))
				})

				Context("when initialising the volume fails", func() {
					BeforeEach(func() {
						initVolume.InitializeReturns(nil, errors.New("Oh gosh, I just don't know what to say..."))
					})

					It("fails with a sensible error message", func() {
						_, err := strategyProvider.ProvideStrategy(rootFsPath)
						Expect(err).To(MatchError("Oh gosh, I just don't know what to say..."))
					})
				})

				It("copies the contents of the rootfs to the volume", func() {
					_, err := strategyProvider.ProvideStrategy(rootFsPath)
					Expect(err).NotTo(HaveOccurred())

					Expect(runner).To(HaveExecutedSerially(fake_command_runner.CommandSpec{
						Path: "cp",
						Args: []string{"-R", rootFsPath, "/path/to/volume"},
					}))
				})

				Context("when copying the rootfs files fails", func() {
					BeforeEach(func() {
						runner.WhenRunning(fake_command_runner.CommandSpec{
							Path: "cp",
						}, func(_ *exec.Cmd) error {
							return errors.New("Ermagahd! Anerther Eeeeeror!")
						})
					})

					It("fails with a sensible error message", func() {
						_, err := strategyProvider.ProvideStrategy(rootFsPath)
						Expect(err).To(MatchError("Ermagahd! Anerther Eeeeeror!"))
					})

					It("should not initialise the volume", func() {
						_, err := strategyProvider.ProvideStrategy(rootFsPath)
						Expect(err).To(HaveOccurred())

						Expect(initVolume.InitializeCallCount()).To(Equal(0))
					})

					It("deletes the now-useless volume", func() {

						_, err := strategyProvider.ProvideStrategy(rootFsPath)
						Expect(err).To(HaveOccurred())

						Expect(initVolume.DestroyCallCount()).To(Equal(1))
					})
				})

				It("returns a strategy", func() {
					strategy, err := strategyProvider.ProvideStrategy(rootFsPath)
					Expect(err).NotTo(HaveOccurred())
					Expect(strategy).NotTo(BeNil())
				})
			})
		})
	})
})
