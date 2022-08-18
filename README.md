# PassPol  
  
Get the passwords that adhere to your target policy out of your password dumps and wordlists.  

Install:  

```
go install github.com/crawl3r/passpol@latest
```
  
Expects a wordlist with a single password per new line (see common100.txt for example).  
  
```
lol
lol123
hello$
cantcode
```
  
Usage:  

```
# only get passwords with more than 4 characters
./passpol -f "passwords.txt" -min 4

# only get passwords that are between 5 and 10 characters long
./passpol -f "passwords.txt" -min 5 -max 10

# only get passwords that have more than 2 special characters re: `(?m)([^A-Za-z0-9])`
./passpol -f "passwords.txt" -sp 2
```

Currently dumps a EOF and time to parse to stdout, ignore these. Currently testing with some big files to see how we do. :)
  
--  
  
Big thanks to https://medium.com/swlh/processing-16gb-file-in-seconds-go-lang-3982c235dfa2 for the examples, used these as a starting point.