package bolt_test

import (
	"github.com/fbiville/go-usain-go/pkg/internal/bolt"
	. "github.com/onsi/gomega"
	"io"
	"net"
	"testing"
)

func TestMessageChunking(t *testing.T) {
	RegisterTestingT(t)

	left, right := net.Pipe()
	chunker := &bolt.Chunker{
		Connection: left,
	}

	go func() {
		err := chunker.WriteChunked([]byte{1, 2, 3, 4})

		Expect(err).NotTo(HaveOccurred(), "must write chunk to left side")
		err = left.Close()
		Expect(err).NotTo(HaveOccurred(), "must send EOF to unblock read from right side")
	}()

	chunk, err := io.ReadAll(right)
	Expect(err).NotTo(HaveOccurred())
	Expect(chunk).To(Equal([]byte{0, 4, 1, 2, 3, 4, 0, 0}))
}

func TestMessageUnchunking(t *testing.T) {
	RegisterTestingT(t)

	left, right := net.Pipe()
	chunker := &bolt.Chunker{
		Connection: right,
	}

	go func() {
		_, err := left.Write([]byte{0, 9, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 0})
		Expect(err).NotTo(HaveOccurred(), "must write chunk to left side")
		err = left.Close()
		Expect(err).NotTo(HaveOccurred(), "must send EOF to unblock read from right side")
	}()

	message, err := chunker.ReadUnchunked()

	Expect(err).NotTo(HaveOccurred())
	Expect(message).To(Equal([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9}))
}
