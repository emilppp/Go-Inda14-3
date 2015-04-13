// http://www.nada.kth.se/~snilsson/concurrency/
package main

import (
	"fmt"
	"sync"
)

// This programs demonstrates how a channel can be used for sending and
// receiving by any number of goroutines. It also shows how  the select
// statement can be used to choose one out of several communications.
func main() {
	people := []string{"Anna", "Bob", "Cody", "Dave", "Eva", "Runar"}
	match := make(chan string, 1) // Make room for one unmatched send.
	wg := new(sync.WaitGroup)
	wg.Add(len(people))
	for _, name := range people {
		go Seek(name, match, wg)
	}
	wg.Wait()
	select {
	case name := <-match:
		fmt.Printf("No one received %ss message.\n", name)
	default:
		// There was no pending send operation.
	}
}

// Seek either sends or receives, whichever possible, a name on the match
// channel and notifies the wait group when done.
func Seek(name string, match chan string, wg *sync.WaitGroup) {
	select {
	case peer := <-match:
		fmt.Printf("%s sent a message to %s.\n", peer, name)
	case match <- name:
		// Wait for someone to receive my message.
	}
	wg.Done()
}

/*
FRÅGOR.

Vad händer om man tar bort go-kommandot från Seek-anropet i main-funktionen?
- Inget speciellt händer, hade det dock varit en obuffrad kanal hade det blivit deadlock.

Vad händer om man byter deklarationen wg := new(sync.WaitGroup) mot var wg sync.WaitGroup och parametern wg *sync.WaitGroup mot wg sync.WaitGroup?
- Då skickar man istället med en kopia var (istället för en referens av en och samma) av waitgroupen varje gång, där den då förblir oförändrad.
Det blir deadlock.

Vad händer om man tar bort bufferten på kanalen match?
- Det blir deadlock. Eftersom att kanalen då kan behandla mer än ett element i taget och då försöker läsa från och skriva till samma kanal.

Vad händer om man tar bort default-fallet från case-satsen i main-funktionen?
- Deadlock ifall antal personer är jämnt. Detta pga första name, fältet under select, hela tiden förväntar sig att ett namn skall kunna hämtas
ur kanalen, som är tom med tanke på det jämna antalet namn. Eftersom att inget namn finns där så kommer hela tiden kanalen vänta på att skicka ut,
vilket aldrig sekr, och det blir deadlock. Med default kan select däremot gå vidare och välja ett annat alternativ, att låta kanalen förbli tom och istället
avsluta programmet.
*/
