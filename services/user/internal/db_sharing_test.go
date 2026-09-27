package internal

import (
	"testing"

	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/pkg/dbx"
)

// TestNewServerSharedVsStandalone 验证 DB 复用契约:
//   - 默认(Standalone=false)必须复用宿主 global.DB,且不覆盖 global.DB;
//   - Standalone=true 必须自建独立库,同样不覆盖 global.DB。
func TestNewServerSharedVsStandalone(t *testing.T) {
	shared, err := dbx.Open("sqlite", "file:shared_role?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open shared: %v", err)
	}
	global.DB = shared

	// 共享模式:复用 global.DB
	s1, err := NewServer(Config{Driver: "sqlite", DSN: "ignored"})
	if err != nil {
		t.Fatalf("shared NewServer: %v", err)
	}
	if s1.db != shared {
		t.Fatalf("共享模式应复用 global.DB 指针")
	}
	if global.DB != shared {
		t.Fatalf("共享模式不得覆盖 global.DB")
	}

	// 独立模式:自建连接池,且不覆盖 global.DB
	s2, err := NewServer(Config{Driver: "sqlite", DSN: "file:standalone_role?mode=memory&cache=shared", Standalone: true})
	if err != nil {
		t.Fatalf("standalone NewServer: %v", err)
	}
	if s2.db == shared {
		t.Fatalf("独立模式应使用自己的连接池")
	}
	if global.DB != shared {
		t.Fatalf("独立模式不得覆盖 global.DB")
	}
}
