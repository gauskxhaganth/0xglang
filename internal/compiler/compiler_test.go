package compiler

import (
	"bytes"
	"strings"
	"testing"
)

// Uji Lapis 7: Uji Eksekutor Latar Belakang (Compiler Proxy End-to-End)
func TestCompilerProxyRun(t *testing.T) {
	// Skrip murni 0xg
	src := `cabinet main

require "fmt" retain

def main()
	let result = 15 + 5
	fmt.Println(result)
end`

	var out bytes.Buffer
	// Run script and capture Standard Output
	err := RunSource([]byte(src), &out)
	if err != nil {
		t.Fatalf("RunSource error: %v, out: %s", err, out.String())
	}
}

// Uji Lapis 8: Source Mapping (Compiler Directive //line)
func TestSourceMappingError(t *testing.T) {
	src := `cabinet main
def main()
	let x = "string" + 5
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err == nil {
		t.Fatalf("FASE MERAH: Diharapkan gagal kompilasi (Type Error), tapi sukses")
	}

	errStr := out.String()
	if !strings.Contains(errStr, "main.0xg:3") {
		t.Errorf("FASE MERAH: Pesan error compiler tidak merujuk ke main.0xg:3. Output error: %s", errStr)
	}
}

// Layer 9 Test: Function Call (CallExpr)
func TestCallExpr(t *testing.T) {
	src := `cabinet main
def main()
	let a = 10
	let b = 20
	println(a + b)
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("FASE MERAH: Gagal mengeksekusi CallExpr. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "30") {
		t.Errorf("FASE MERAH: Diharapkan mencetak '30', tapi dapat: '%s'", errStr)
	}
}

// Uji Lapis 10: Interoperabilitas Library (RequireDecl & SelectorExpr)
func TestRequireStmt(t *testing.T) {
	src := `cabinet main

require "fmt" retain

def main()
	fmt.Println("Halo dari 0xg!")
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("FASE MERAH: Gagal mengeksekusi RequireDecl. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "Halo dari 0xg!") {
		t.Errorf("FASE MERAH: Diharapkan mencetak 'Halo dari 0xg!', tapi dapat: '%s'", errStr)
	}
}

// Uji Lapis 11: Pengembalian Nilai (ReturnStmt)
func TestReturnStmt(t *testing.T) {
	src := `cabinet main
require "fmt" retain

def Hitung(a Int, b Int) Int
	return a * b
end

def main()
	let hasil = Hitung(5, 5)
	fmt.Println(hasil)
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("FASE MERAH: Gagal mengeksekusi ReturnStmt. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "25") {
		t.Errorf("FASE MERAH: Diharapkan mencetak '25', tapi dapat: '%s'", errStr)
	}
}

// Uji Lapis 12: Deklarasi Objek (Struct)
func TestStructDecl(t *testing.T) {
	src := `cabinet main
require "fmt" retain

struct Pengguna
	nama String
	umur Int
end

def main()
	let user Pengguna
	user.nama = "Budi"
	user.umur = 30
	fmt.Println(user.nama, user.umur)
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("FASE MERAH: Gagal mengeksekusi Struct. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "Budi 30") {
		t.Errorf("FASE MERAH: Diharapkan mencetak 'Budi 30', tapi dapat: '%s'", errStr)
	}
}

// Uji Lapis 13: Struktur Data Dinamis (Array & Map)
func TestDataStructs(t *testing.T) {
	src := `cabinet main
require "fmt" retain

def main()
	let arr = Array(String){"A", "B", "C"}
	let kamus = Map(String, Int){"Umur": 30}
	
	arr[0] = "Z"
	kamus["Skor"] = 100

	fmt.Println(arr[0], kamus["Umur"], kamus["Skor"])
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("FASE MERAH: Gagal mengeksekusi Array/Map. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "Z 30 100") {
		t.Errorf("FASE MERAH: Diharapkan mencetak 'Z 30 100', tapi dapat: '%s'", errStr)
	}
}

// Uji Lapis 14: Iterasi Kolektif (loop & foreach)
func TestLoops(t *testing.T) {
	src := `cabinet main
require "fmt" retain

def main()
	let arr = Array(String){"Satu", "Dua"}
	
	foreach idx, val in arr
		fmt.Println(idx, val)
	end
	
	let k = 0
	loop
		if k == 2
			break
		end
		fmt.Println("Loop", k)
		k = k + 1
	end
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("FASE MERAH: Gagal mengeksekusi loop/foreach. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "0 Satu") || !strings.Contains(errStr, "1 Dua") || !strings.Contains(errStr, "Loop 0") || !strings.Contains(errStr, "Loop 1") {
		t.Errorf("FASE MERAH: Output iterasi tidak sesuai, didapat: '%s'", errStr)
	}
}

// Uji Lapis 15: Percabangan Banyak Opsi (case & when)
func TestSwitchCase(t *testing.T) {
	src := `cabinet main
require "fmt" retain

def CekAngka(x Int)
	case x
	when 1
		fmt.Println("Satu")
	when 2, 3
		fmt.Println("DuaTiga")
	else
		fmt.Println("Lainnya")
	end
end

def main()
	CekAngka(1)
	CekAngka(3)
	CekAngka(99)
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("FASE MERAH: Gagal mengeksekusi case/when. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "Satu") || !strings.Contains(errStr, "DuaTiga") || !strings.Contains(errStr, "Lainnya") {
		t.Errorf("FASE MERAH: Output switch-case tidak sesuai, didapat: '%s'", errStr)
	}
}

// Uji Lapis 16: Error Handling (Blood Lock Metaprogramming)
func TestErrorHandling(t *testing.T) {
	src := `cabinet main
require "fmt" retain
require "errors" retain

def CekError(pemicu Int)
	let err = errors.New("Bahaya!")
	
	if pemicu == 1
		err = nil
	end

	if err
		fmt.Println("Gagal:", err.Error())
	end
	
	if !err
		fmt.Println("Sukses!")
	end
end

def main()
	CekError(0)
	CekError(1)
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("FASE MERAH: Gagal mengeksekusi error handling. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "Gagal: Bahaya!") || !strings.Contains(errStr, "Sukses!") {
		t.Errorf("FASE MERAH: Output if err tidak sesuai, didapat: '%s'", errStr)
	}
}

// Uji Lapis 17: Konkurensi & Memory (go, defer, Channel)
func TestConcurrency(t *testing.T) {
	src := `cabinet main
require "fmt" retain
require "time" retain

def Pekerja(id Int, c Channel(String))
	time.Sleep(10 * time.Millisecond)
	c <- "Tugas Selesai"
end

def main()
	let c = make(Channel(String))
	defer fmt.Println("Defer Dieksekusi")
	
	go Pekerja(1, c)
	
	let hasil = <-c
	fmt.Println("Pesan:", hasil)
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("FASE MERAH: Gagal mengeksekusi konkurensi. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	// Karena defer dieksekusi di akhir (setelah fungsi selesai), pesan dicetak sebelum/setelah bergantung pada flush, 
	// namun di Go murni output standar akan menangkap keduanya.
	if !strings.Contains(errStr, "Pesan: Tugas Selesai") || !strings.Contains(errStr, "Defer Dieksekusi") {
		t.Errorf("FASE MERAH: Output konkurensi tidak lengkap, didapat: '%s'", errStr)
	}
}

// Uji Lapis 18: Konkurensi Lanjutan (Select)
func TestSelectStatement(t *testing.T) {
	src := `cabinet main
require "fmt" retain
require "time" retain

def Kirim1(c Channel(String))
	time.Sleep(10 * time.Millisecond)
	c <- "Satu"
end

def Kirim2(c Channel(String))
	time.Sleep(20 * time.Millisecond)
	c <- "Dua"
end

def main()
	let c1 = make(Channel(String))
	let c2 = make(Channel(String))

	go Kirim1(c1)
	go Kirim2(c2)

	select
	when <-c1
		fmt.Println("Dapat dari c1: Satu")
	when <-c2
		fmt.Println("Dapat dari c2: Dua")
	when <-time.After(50 * time.Millisecond)
		fmt.Println("Timeout")
	end
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("FASE MERAH: Gagal mengeksekusi select. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "Dapat dari c1: Satu") {
		t.Errorf("FASE MERAH: Output select tidak sesuai, didapat: '%s'", errStr)
	}
}
