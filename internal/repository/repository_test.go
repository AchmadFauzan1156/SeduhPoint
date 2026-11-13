package repository

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestBisaDiulang(t *testing.T) {
	cases := []struct {
		nama string
		err  error
		mau  bool
	}{
		{"duplicate entry (token sama)", &mysql.MySQLError{Number: 1062}, true},
		{"deadlock", fmt.Errorf("tambah transaksi: %w", &mysql.MySQLError{Number: 1213}), true},
		{"lock wait timeout", &mysql.MySQLError{Number: 1205}, true},
		{"kolom tidak dikenal", &mysql.MySQLError{Number: 1054}, false},
		{"error biasa", errors.New("koneksi putus"), false},
		{"tanpa error", nil, false},
	}
	for _, c := range cases {
		if got := BisaDiulang(c.err); got != c.mau {
			t.Errorf("%s: dapat %v, harusnya %v", c.nama, got, c.mau)
		}
	}
}
