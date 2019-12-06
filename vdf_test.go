package main

import (
	cr "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"
)

func TestProof(t *testing.T) {
	max := new(big.Int).SetUint64(1 << 63)
	seed, err := cr.Int(cr.Reader, max)
	if err != nil {
		t.Error("random failed")
	}
	rand.Seed(seed.Int64())
	for i := 0; i < 4; i++ {
		security := uint64(512)
		N := Setup(security)
		instance := Generate(N, 256, security)
		instance.NaiveSolve()
		if !instance.Verify() {
			t.Error("verification failed")
		}
		//lets flip some bits in the proof
		for i := 0; i < 100; i++ {
			el := rand.Intn((len(instance.mu)))

			bytess := instance.mu[el].Bytes()
			old := new(big.Int).Set(instance.mu[el])
			bitpos := rand.Intn(len(bytess) * 8)
			bitToflip := old.Bit(bitpos)
			var newer *big.Int
			if bitToflip == 1 {
				newer = new(big.Int).SetBit(old, bitpos, 0)
			} else {
				newer = new(big.Int).SetBit(old, bitpos, 1)
			}
			instance.mu[el] = newer
			if instance.Verify() {
				t.Error("bit flip must cause verify to reject with overwhelming prob")
			}
			instance.mu[el] = old
			if !instance.Verify() {
				t.Error("verification failed")
			}
			//fmt.Println(newer.String())
			//fmt.Println(old.String())
		}

	}

}

//TODO remove
func BenchmarkSetup(b *testing.B) {
	for i := 1 << 9; i < 1<<62; i = i << 1 {
		fmt.Println(i)
		for j := 0; j < 10; j++ {
			security := uint64(i)
			p, _ := cr.Prime(cr.Reader, int(security/2))
			q, _ := cr.Prime(cr.Reader, int(security/2))
			N := new(big.Int).Mul(p, q)
			var x, _ = cr.Int(cr.Reader, N)
			before := time.Now()
			a := new(big.Int).GCD(nil, nil, x, N)
			fmt.Println("gcd:", time.Since(before), " at", i)
			a.String()
			before = time.Now()
			big.Jacobi(x, N)
			fmt.Println("J:", time.Since(before), " at", i)
		}
	}
}

func TestSquare(t *testing.T) {
	r := new(big.Int).SetInt64(rand.Int63())
	mod, _ := cr.Prime(cr.Reader, int(256))
	T := uint64(1 << uint64(rand.Int31()))
	mu1 := Square(r, mod, T/2)
	y1 := Square(mu1, mod, T/2)
	if Square(r, mod, T).Cmp(y1) != 0 {
		t.Error("unequal")
	}
}

func TestIsPowerTwo(t *testing.T) {
	IsPowerTwo(4)

	for i := 0; i < 10000; i++ {
		if n := rand.Uint64(); IsPowerTwo(n) {
			fmt.Println(n)
		}
	}
	for i := uint64(0); i < 64; i++ {
		if n := (1 << i); IsPowerTwo(uint64(n)) {
			fmt.Println(i, n)
		}
	}
}
