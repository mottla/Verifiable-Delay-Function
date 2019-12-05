package main

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"math/big"
	"sync"
)

var bigOne = new(big.Int).SetInt64(int64(1))
var bigTwo = new(big.Int).SetInt64(int64(2))

//Setup takes a security parameter and outputs the product of two safe primes
func Setup(security uint64) *big.Int {

	var err error

	chk := func(dest *big.Int, wg *sync.WaitGroup) {
		ctr := 0
		var tmp *big.Int
		for {
			ctr++
			tmp, err = rand.Prime(rand.Reader, int(security/2))

			if err != nil {
				panic(err)
			}
			if isSavePrime(tmp) {
				break
			}
		}
		fmt.Printf("\ntook %v attempts to generate a save prime", ctr)
		dest.Set(tmp)
		wg.Done()
	}
	var p, q = new(big.Int), new(big.Int)
	//lets find the primes concurrently, cuz we can. its golang bliad
	wg := sync.WaitGroup{}
	wg.Add(2)
	go chk(p, &wg)
	go chk(q, &wg)
	wg.Wait()

	rsaModulus := new(big.Int).Mul(p, q)
	return rsaModulus

}

//isSavePrime takes a prime p and checks if (p-1)/2 is a prime then outputs true
func isSavePrime(prime *big.Int) bool {
	tmp := new(big.Int).Sub(prime, bigOne)
	tmp.Div(tmp, bigTwo)
	return tmp.ProbablyPrime(20)
}

//Generate takes a rsaModulus of two safe primes, a security parameter and the squaring parameter T
//it computes a challange x, which is elment of the quadratic residue class QR+ whose membership can be efficiently checked
func Generate(rsaModulus *big.Int, T, security uint64) Instance {

	if !IsPowerTwo(T) {
		panic("currently time parameter only powers of two allowed ")
	}

	//generate random x in [0,N)
	var x, err = rand.Int(rand.Reader, rsaModulus)
	//check if gcd(x,N)=1, to ensure that x in Z_N*
	for big.Jacobi(x, rsaModulus) == 0 {
		x, err = rand.Int(rand.Reader, rsaModulus)
	}
	if err != nil {
		panic(err)
	}
	//set x <- x²|N to ensure its a member of the quadratic residue class
	x.Mul(x, x)
	x.Mod(x, rsaModulus)

	return Instance{
		rsaModulus:   rsaModulus,
		challenge:    x,
		T:            T,
		y:            nil,
		mu:           nil,
		securityBits: security,
	}

}

type Instance struct {
	rsaModulus, challenge *big.Int
	T                     uint64
	y                     *big.Int
	mu                    []*big.Int
	securityBits          uint64
}

//NaiveSolve runs on a instance that has been instanciated via Generate before
//it solves the RSW timelock puzzle and stores the proof instance in the calling instantiation
//this implementation is naive, as it does not performe the suggested optimiced solve algorithm in the paper.
func (in *Instance) NaiveSolve() {
	mu1 := Square(in.challenge, in.rsaModulus, in.T/2)
	y1 := Square(mu1, in.rsaModulus, in.T/2)
	in.y = y1
	r1 := in.hash(in.challenge, y1, mu1, in.T)
	in.mu = append(in.mu, mu1)
	xii := new(big.Int).Exp(in.challenge, r1, in.rsaModulus)
	xii.Mul(xii, mu1)
	xii.Mod(xii, in.rsaModulus)

	yii := new(big.Int).Exp(mu1, r1, in.rsaModulus)
	yii.Mul(yii, y1)
	yii.Mod(yii, in.rsaModulus)
	in.naiveSolve(xii, yii, 1<<2)
}

func (in *Instance) naiveSolve(xi, yi *big.Int, twoPoweri uint64) {
	if twoPoweri > in.T {
		return
	}
	mui := Square(xi, in.rsaModulus, in.T/(twoPoweri))
	in.mu = append(in.mu, mui)
	ri := in.hash(xi, yi, mui, in.T/(twoPoweri>>1))

	xii := new(big.Int).Exp(xi, ri, in.rsaModulus)
	xii.Mul(xii, mui)
	xii.Mod(xii, in.rsaModulus)

	yii := new(big.Int).Exp(mui, ri, in.rsaModulus)
	yii.Mul(yii, yi)
	yii.Mod(yii, in.rsaModulus)
	in.naiveSolve(xii, yii, twoPoweri<<1)
}

//Verify called on a solved instance checks, if the VDF has been computed correctly
func (in *Instance) Verify() bool {
	x := in.challenge
	y := in.y
	var ri *big.Int
	for i := uint64(0); i < uint64(len(in.mu)); i++ {
		ri = in.hash(x, y, in.mu[i], in.T/(1<<i))
		xii := new(big.Int).Exp(x, ri, in.rsaModulus)
		xii.Mul(xii, in.mu[i])
		xii.Mod(xii, in.rsaModulus)
		x = xii
		yii := new(big.Int).Exp(in.mu[i], ri, in.rsaModulus)
		yii.Mul(yii, y)
		yii.Mod(yii, in.rsaModulus)
		y = yii
	}
	return Square(x, in.rsaModulus, 1).Cmp(y) == 0
}

// start in -> in²^target
func Square(in, mod *big.Int, target uint64) (res *big.Int) {
	res = new(big.Int).Set(in)
	return square(res, mod, 0, target)
}

func square(in, mod *big.Int, squareTimes, target uint64) *big.Int {
	if squareTimes == target {
		return in
	}
	squareTimes++
	in.Mul(in, in)
	return square(in.Mod(in, mod), mod, squareTimes, target)
}

//note that this is a sloppy solution. If bitsize of the rsa modulus is bigger then 512, we may lose something.. however
//i do not understand how problematic that would be
func (in *Instance) hash(x, y, mu *big.Int, T uint64) *big.Int {
	b := sha512.New()
	b.Write(x.Bytes())
	b.Write(y.Bytes())
	b.Write(mu.Bytes())
	bits := make([]byte, 8)
	binary.LittleEndian.PutUint64(bits, T)
	b.Write(bits)
	res := new(big.Int).SetBytes(b.Sum(nil))
	res.Mod(res, in.rsaModulus)
	return res
}

//checks if a integer has only one 1 bit
func IsPowerTwo(in uint64) bool {
	ctr := uint64(0)
	for i := uint64(0); i < 64; i++ {
		ctr += ((in & (1 << i)) >> i)
	}
	if ctr == 1 {
		return true
	}
	return false
}
