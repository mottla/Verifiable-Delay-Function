# Verifiable Delay Funciton 
Verifiable Delay Functions implementation in Go, based on


- `Simple Verifiable Delay Functions`, Krzysztof Pietrzak https://research-explorer.app.ist.ac.at/download/6528/6529/2019_LIPIcs_Pietrzak.pdf


## Caution & Warning
Its crypto. What can possibly go wrong?!

##Usage

To setup a VDF with 512 bit RSA security
```
security := uint64(512)
N := Setup(security)

```

To create a lock that requires min 256 steps
```
instance := Generate(N, 256, security)
```

Solve the puzzle. Currently we only implement the naive solver
```
instance.NaiveSolve()
```
To verify an instance, call and check if true/false
```
accept := instance.Verify() 
	
```