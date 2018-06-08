package bash_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"testing"
	"unicode"

	"github.com/progrium/go-basher"

	"github.com/onsi/gomega/gbytes"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/jhvhs/gob-mock"
	. "github.com/onsi/gomega"
)

func makeTitleCase(args []string) {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	runes := []rune(strings.Trim(string(bytes), "\n"))
	spacePrev := false
	for i := range runes {

		switch {
		case unicode.IsLetter(runes[i]) && (i == 0 || spacePrev):
			runes[i] = unicode.ToUpper(runes[i])
			spacePrev = false
		case unicode.IsSpace(runes[i]):
			spacePrev = true
		default:
			spacePrev = false
		}
	}

	fmt.Println(string(runes))
}

func TestBashFunctions(t *testing.T) {
	spec.Run(t, "Bash Script Tests", func(t *testing.T, when spec.G, it spec.S) {
		var bash *basher.Context
		it.Before(func() {
			RegisterTestingT(t)

			var err error
			bash, err = basher.NewContext(getInternalBash(), false)
			Expect(err).NotTo(HaveOccurred())

			err = bash.Source("../scripts/script-under-test.bash", nil)
			Expect(err).NotTo(HaveOccurred())

			bash.CopyEnv()

			bash.Stdout = gbytes.NewBuffer()
			bash.Stderr = gbytes.NewBuffer()
		})

		when("Testing pure bash functions", func() {
			it("returns correctly", func() {
				code, err := bash.Run("capitalize", []string{"hello", "world"})
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(BeZero())
				Expect(bash.Stdout).To(gbytes.Say("HELLO WORLD"))
				Expect(bash.Stderr).To(gbytes.Say(""))

			})
		})

		when("Using a go func in Bash", func() {
			it("runs correctly", func() {
				bash.ExportFunc("make-title-case", makeTitleCase)
				if bash.HandleFuncs(os.Args) {
					os.Exit(0)
				}
				code, err := bash.Run("convertToTitle", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(BeZero())
				Expect(bash.Stdout).To(gbytes.Say("This Should Be In Title Case"))
			})
		})

		when("Testing a command that fails", func() {
			it("fails", func() {
				code, err := bash.Run("testStub", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(code).NotTo(BeZero())
			})

			it("works when stubbed", func() {
				mocks := []Gob{Stub("false")}
				ApplyMocks(bash, mocks)

				code, err := bash.Run("testStub", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(BeZero())
			})
		})

		when("Testing with spies", func() {
			it("returns correctly with info", func() {
				mocks := []Gob{Spy("cf")}
				ApplyMocks(bash, mocks)

				code, err := bash.Run("testSpies", []string{"-asdf", "this would fail if cf weren't stubbed"})
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(BeZero())
				Expect(bash.Stdout).To(gbytes.Say(""))
				Expect(bash.Stderr).To(gbytes.Say("<1> cf version"))
				Expect(bash.Stderr).To(gbytes.Say("<2> cf push -asdf this would fail if cf weren't stubbed"))
			})

			it("falls through when using SpyAndConditionallyCallThrough", func() {
				mocks := []Gob{SpyAndConditionallyCallThrough("cf", `[ $1 == 'version' ]`)}
				ApplyMocks(bash, mocks)

				code, err := bash.Run("testSpies", []string{"-asdf", "this would fail if cf weren't stubbed"})
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(BeZero())
				Expect(bash.Stdout).To(gbytes.Say("cf version \\d+\\.\\d+\\.\\d+\\+.*"))
				Expect(bash.Stderr).To(gbytes.Say("<1> cf version"))
				Expect(bash.Stderr).To(gbytes.Say("<2> cf push -asdf this would fail if cf weren't stubbed"))
			})

			it("falls through with stdin too", func() {
				input := "input ... from the fuuuture!\n"
				bash.Stdin = bytes.NewBuffer([]byte(input))
				mocks := []Gob{SpyAndCallThrough("cat")}

				ApplyMocks(bash, mocks)
				code, err := bash.Run("readInput", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(BeZero())
				Expect(bash.Stdout).To(gbytes.Say(input))
				Expect(bash.Stderr).To(gbytes.Say("<1> cat"))
			})
		})

		when("Testing with mocks", func() {
			it("works with all types of mocks", func() {
				mocks := []Gob{
					Mock("cf", "echo 'No PCF for you bye bye'"),
					MockOrCallThrough("bosh", "echo 'hello'", `[ $1 == '--version' ]`),
				}
				ApplyMocks(bash, mocks)

				code, err := bash.Run("testMocks", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(BeZero())
				Expect(bash.Stdout).To(gbytes.Say("No PCF for you bye bye"))
				Expect(bash.Stdout).To(gbytes.Say("version \\d+\\.\\d+\\.\\d+\\-.*"))
				Expect(bash.Stdout).To(gbytes.Say("hello"))
				Expect(bash.Stderr).To(gbytes.Say("<1> cf help"))
				Expect(bash.Stderr).To(gbytes.Say("<2> bosh --version"))
				Expect(bash.Stderr).To(gbytes.Say("<3> bosh envs"))
			})
		})
	}, spec.Report(report.Terminal{}))
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
