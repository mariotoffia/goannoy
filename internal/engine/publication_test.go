package engine

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/mariotoffia/goannoy/internal/codec"
)

func TestInvalidBuildNeverPublishes(t *testing.T) {
	for _, disk := range []bool{false, true} {
		t.Run(map[bool]string{false: "memory", true: "on-disk"}[disk], func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "index.ann")
			b, _ := NewBuilder(codec.Angular, 1)
			b.Config.Workers(1).Trees(1)
			if err := b.AddFloat(0, []float32{1}); err != nil {
				t.Fatal(err)
			}
			old, err := b.BuildTo(nil, path)
			if err != nil {
				t.Fatal(err)
			}
			defer old.Close()
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			// Force a bad derived node independently of boundary validation.
			codec.PutF32(b.items, 4, float32(math.Inf(1)))
			if disk {
				b.Config.OnDisk(path)
			}
			if next, err := b.BuildTo(nil, path); err == nil {
				next.Close()
				t.Fatal("published invalid build")
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("failed build replaced old file: %v", err)
			}
			reopened, err := Open(path, codec.Angular, 1)
			if err != nil {
				t.Fatal(err)
			}
			reopened.Close()
			files, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".goannoy-*"))
			if len(files) != 0 {
				t.Fatal("temporary files leaked", files)
			}
		})
	}
}
