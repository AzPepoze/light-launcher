package downloads

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/logger"
)

// decryptChunkSize is how much of the file is decrypted per read.
const decryptChunkSize = 64 * 1024

// ErrDecrypt means a finished download could not be decrypted. The download fails and is not installed.
var ErrDecrypt = errors.New("could not decrypt the download")

// placeFinished puts a finished part file at dest. Without a decrypt spec it is a plain rename.
// With one, the decrypted copy goes to a temp file beside dest and is renamed into place, so dest
// never holds a half-decrypted file. When requireArchive is set, the decrypted header must be an
// archive the installer supports; a wrong key is otherwise undetectable in AES-CTR. The part file
// is left for the caller to remove.
func placeFinished(ctx context.Context, part, dest string, spec *contract.DecryptSpec, requireArchive bool) error {
	if spec == nil {
		return os.Rename(part, dest)
	}
	return decryptIntoPlace(ctx, part, dest, spec, requireArchive)
}

// decryptIntoPlace decrypts src with AES-128-CTR into a temp file and renames it to dest.
// AES-CTR keeps the byte count, so the result has exactly the size of the source file.
func decryptIntoPlace(ctx context.Context, src, dest string, spec *contract.DecryptSpec, requireArchive bool) error {
	key, nonce, err := spec.Decode()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDecrypt, err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDecrypt, err)
	}
	// The 8-byte nonce is the high half of the 16-byte counter block; the counter starts at zero.
	iv := make([]byte, aes.BlockSize)
	copy(iv, nonce)

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDecrypt, err)
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDecrypt, err)
	}
	size := info.Size()

	tmp, err := os.CreateTemp(filepath.Dir(dest), "."+filepath.Base(dest)+".decrypt-*")
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDecrypt, err)
	}
	tmpName := tmp.Name()
	// A successful rename removes tmpName, so this only cleans up a failed decrypt.
	defer os.Remove(tmpName)
	// Match the mode a plain downloaded file gets.
	_ = tmp.Chmod(0o644)

	written, copyErr := decryptCopy(ctx, tmp, in, cipher.NewCTR(block, iv))
	closeErr := tmp.Close()
	switch {
	case copyErr != nil:
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("%w: %v", ErrDecrypt, copyErr)
	case closeErr != nil:
		return fmt.Errorf("%w: %v", ErrDecrypt, closeErr)
	case written != size:
		return fmt.Errorf("%w: decrypted %d of %d bytes", ErrDecrypt, written, size)
	}
	if requireArchive && kindByMagic(tmpName) == KindUnknown {
		return fmt.Errorf("%w: the decrypted file is not an archive this launcher can install", ErrDecrypt)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		return fmt.Errorf("%w: %v", ErrDecrypt, err)
	}
	return nil
}

// decryptCopy writes the CTR-decrypted copy of src to dst, stopping when the context ends.
func decryptCopy(ctx context.Context, dst io.Writer, src io.Reader, stream cipher.Stream) (int64, error) {
	buf := make([]byte, decryptChunkSize)
	var written int64
	for {
		if err := ctx.Err(); err != nil {
			return written, err
		}
		n, readErr := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf[:n], buf[:n])
			if _, err := dst.Write(buf[:n]); err != nil {
				return written, err
			}
			written += int64(n)
		}
		if errors.Is(readErr, io.EOF) {
			return written, nil
		}
		if readErr != nil {
			return written, readErr
		}
	}
}

// removePart drops the encrypted part after the decrypted copy reached dest. A missing file is fine.
func removePart(part string) {
	if err := os.Remove(part); err != nil && !os.IsNotExist(err) {
		logger.Warn(logCategory, "could not delete %s: %v", part, err)
	}
}
