package mindconnect_test

import (
	"fmt"

	"github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func ExampleSplitBatches() {
	reqs := []mindconnect.Request{
		mindconnect.Var("hp"),
		mindconnect.Var("draw.GetSymbol", mindconnect.Int(0), mindconnect.Int(0)),
		mindconnect.Cmd("equip", "sword"),
	}
	batches, err := mindconnect.SplitBatches(reqs, mindconnect.SafePacket)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i, batch := range batches {
		fmt.Println(i+1, batch)
	}
}

func ExampleEncodeBatch() {
	reqs := []mindconnect.Request{mindconnect.Var("hp"), mindconnect.Var("maxhp")}
	packet, reader, err := mindconnect.EncodeBatch(1, reqs, mindconnect.SafePacket)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%q\n", packet)
	if _, err := reader.Feed([]byte("1\x1f40\x1f50\x1fTrue")); err != nil {
		fmt.Println(err)
		return
	}
	_, values, err := reader.Response()
	fmt.Println(values, err)
}

func ExampleResponseReader() {
	reader := mindconnect.NewResponseReader(7, 3)
	for _, part := range []string{"7\x1f4", "0\x1f50\x1fTr", "ue"} {
		done, err := reader.Feed([]byte(part))
		if err != nil {
			fmt.Println(err)
			return
		}
		if done {
			break
		}
	}
	id, values, err := reader.Response()
	fmt.Println(id, values, err)
}
