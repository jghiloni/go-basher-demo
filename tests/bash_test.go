package bash_test

import (
	"log"
	"os"
	"path"
	"testing"

	"github.com/progrium/go-basher"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

func TestBashFunctions(t *testing.T) {
	spec.Run(t, "Credhub Client", func(t *testing.T, when spec.G, it spec.S) {
		var bash *basher.Context
		it.Before(func() {
			RegisterTestingT(t)

			var err error
			bash, err = basher.NewContext(getInternalBash(), true)
			Expect(err).NotTo(HaveOccurred())
		})

		// bash, _ := basher

		when("Testing bash functions", func() {
			it.Before(func() {
				err := bash.Source("../scripts/script-under-test.bash", nil)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns correctly", func() {
				stdout := gbytes.NewBuffer()
				stderr := gbytes.NewBuffer()

				bash.Stdout = stdout
				bash.Stderr = stderr

				code, err := bash.Run("capitalize", []string{"hello", "world"})
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(BeZero())
				Expect(stdout).To(gbytes.Say("HELLO WORLD"))
				Expect(stderr).To(gbytes.Say(""))

				code, err = bash.Run("capitalize", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(BeZero())
				Expect(stdout).To(gbytes.Say(""))
				Expect(stderr).To(gbytes.Say(""))
			})
		})
	})
}

func getInternalBash() string {
	bashDir := path.Join(os.Getenv("HOME"), ".basher")

	bashPath := bashDir + "/bash"
	if _, err := os.Stat(bashPath); os.IsNotExist(err) {
		err = basher.RestoreAsset(bashDir, "bash")
		if err != nil {
			log.Fatal(err, "1")
		}
	}

	return bashPath
}
