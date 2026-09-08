package clientmanager

import (
	"strconv"
	"testing"

	"github.com/luthermonson/go-proxmox"
)

func benchmarkInternalClient() *proxmox.Client {
	return &proxmox.Client{}
}

func BenchmarkClientManager_get(b *testing.B) {
	cm := New()
	client := benchmarkInternalClient()

	cm.set("primary", client)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if got := cm.get("primary"); got != client {
			b.Fatal("get returned a different client")
		}
	}
}

func BenchmarkClientManager_get_Missing(b *testing.B) {
	cm := New()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if got := cm.get("missing"); got != nil {
			b.Fatal("get returned a client for a missing name")
		}
	}
}

func BenchmarkClientManager_set(b *testing.B) {
	cm := New()
	client := benchmarkInternalClient()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cm.set("primary", client)
	}
}

func BenchmarkClientManager_set_DifferentNames(b *testing.B) {
	cm := New()
	client := benchmarkInternalClient()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cm.set("client-"+strconv.Itoa(i), client)
	}
}

func BenchmarkClientManager_add(b *testing.B) {
	cm := New()
	client := benchmarkInternalClient()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if ok := cm.add("client-"+strconv.Itoa(i), client); !ok {
			b.Fatalf("add returned false for client %d", i)
		}
	}
}

func BenchmarkClientManager_add_Duplicate(b *testing.B) {
	cm := New()
	first := benchmarkInternalClient()
	second := benchmarkInternalClient()

	if ok := cm.add("primary", first); !ok {
		b.Fatal("initial add returned false")
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if ok := cm.add("primary", second); ok {
			b.Fatal("add returned true for duplicate name")
		}
	}
}

func BenchmarkClientManager_get_Parallel(b *testing.B) {
	cm := New()
	client := benchmarkInternalClient()

	cm.set("primary", client)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if got := cm.get("primary"); got != client {
				b.Fatal("get returned a different client")
			}
		}
	})
}

func BenchmarkClientManager_set_Parallel(b *testing.B) {
	cm := New()
	client := benchmarkInternalClient()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cm.set("primary", client)
		}
	})
}

func BenchmarkClientManager_add_Parallel(b *testing.B) {
	cm := New()
	client := benchmarkInternalClient()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for pb.Next() {
				cm.add("primary", client)
			}
		}
	})
}
