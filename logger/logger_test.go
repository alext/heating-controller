package logger

import (
	"bytes"
	"log"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOutput(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logger")
}

var _ = Describe("logging at various levels", func() {
	var (
		output *bytes.Buffer
	)

	BeforeEach(func() {
		output = new(bytes.Buffer)
		SetDestination(output)

		log.SetFlags(0) // Disable timestamps to make testing easier
	})

	Context("with level set to DEBUG", func() {
		BeforeEach(func() {
			Level = DEBUG
		})

		It("should write debug level messages", func() {
			Debug("string1", "string2", "string3")
			Expect(output.ReadString(byte('\n'))).To(Equal("string1 string2 string3\n"))
			Debugf("int: %d, string: %s", 4, "foo")
			Expect(output.ReadString(byte('\n'))).To(Equal("int: 4, string: foo\n"))
		})

		It("should write info level messages", func() {
			Info("string1", "string2", "string3")
			Expect(output.ReadString(byte('\n'))).To(Equal("string1 string2 string3\n"))
			Infof("int: %d, string: %s", 4, "foo")
			Expect(output.ReadString(byte('\n'))).To(Equal("int: 4, string: foo\n"))
		})

		It("should write warn level messages", func() {
			Warn("string1", "string2", "string3")
			Expect(output.ReadString(byte('\n'))).To(Equal("string1 string2 string3\n"))
			Warnf("int: %d, string: %s", 4, "foo")
			Expect(output.ReadString(byte('\n'))).To(Equal("int: 4, string: foo\n"))
		})
	})

	Context("with level set to INFO", func() {
		BeforeEach(func() {
			Level = INFO
		})

		It("should write nothing for debug level messages", func() {
			Debug("message")
			Expect(output.Len()).To(Equal(0))

			Debugf("message %d", 4)
			Expect(output.Len()).To(Equal(0))
		})

		It("should write info level messages", func() {
			Info("string1", "string2", "string3")
			Expect(output.ReadString(byte('\n'))).To(Equal("string1 string2 string3\n"))
			Infof("int: %d, string: %s", 4, "foo")
			Expect(output.ReadString(byte('\n'))).To(Equal("int: 4, string: foo\n"))
		})

		It("should write warn level messages", func() {
			Warn("string1", "string2", "string3")
			Expect(output.ReadString(byte('\n'))).To(Equal("string1 string2 string3\n"))
			Warnf("int: %d, string: %s", 4, "foo")
			Expect(output.ReadString(byte('\n'))).To(Equal("int: 4, string: foo\n"))
		})
	})

	Context("with level set to WARN", func() {
		BeforeEach(func() {
			Level = WARN
		})

		It("should write nothing for debug level messages", func() {
			Debug("message")
			Expect(output.Len()).To(Equal(0))

			Debugf("message %d", 4)
			Expect(output.Len()).To(Equal(0))
		})

		It("should write nothing for info level messages", func() {
			Info("message")
			Expect(output.Len()).To(Equal(0))

			Infof("message %d", 4)
			Expect(output.Len()).To(Equal(0))
		})

		It("should write warn level messages", func() {
			Warn("string1", "string2", "string3")
			Expect(output.ReadString(byte('\n'))).To(Equal("string1 string2 string3\n"))
			Warnf("int: %d, string: %s", 4, "foo")
			Expect(output.ReadString(byte('\n'))).To(Equal("int: 4, string: foo\n"))
		})
	})
})
