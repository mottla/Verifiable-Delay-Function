# Verifiable Delay Funciton 
Verifiable Delay Functions implementation in Go, based on


- `Simple Verifiable Delay Functions`, Krzysztof Pietrzak https://research-explorer.app.ist.ac.at/download/6528/6529/2019_LIPIcs_Pietrzak.pdf


## Caution & Warning
Its crypto. What can possibly go wrong?!


## Usage
To setup a VDF with k bit RSA security
```
security := uint64(k)
N := Setup(security)

```

To create a lock that requires min T squarings
```
instance := Generate(N, T, k)
```

Solve the puzzle. Currently we only implement the naive solver
```
instance.NaiveSolve()
```
To verify an instance, call and check if true/false
```
accept := instance.Verify() 
	
```