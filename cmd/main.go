package main

import mantydns "manty-dns"

func main() {
	mantydns.Start(10054, "0.0.0.0")
}
