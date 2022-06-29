package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type commandLog struct {
	command string
	args    []string
}

func (cL commandLog) String() string {
	return fmt.Sprintf("{ command: %v, args: %v }", cL.command, strings.Join(cL.args, ", "))
}

type transaction struct {
	parent   *transaction
	commands []commandLog
}

func (t transaction) String() string {
	return fmt.Sprintf("{ parent: { %v }, commands: %v }", t.parent, t.commands)
}

type storeApp struct {
	store             map[string]string
	transactions      []transaction
	activeTransaction *transaction
}

var root = storeApp{
	store:             make(map[string]string),
	transactions:      []transaction{},
	activeTransaction: nil,
}

func printUnknownCommand(command string) {
	fmt.Println("> unknown command: ", command)
}

func readCommand(r *bufio.Reader) string {
	fmt.Print("> ")
	t, _ := r.ReadString('\n')
	return strings.TrimSpace(t)
}

func recordCommand(command string, args []string) {
	root.activeTransaction.commands = append(root.activeTransaction.commands, commandLog{command, args})
}

func shouldContinue(command string) bool {
	return !strings.EqualFold("QUIT", command)
}

func readFromStore(args ...string) {
	key := args[0]

	if value, exists := root.store[key]; exists {
		fmt.Println(value)

		if root.activeTransaction != nil {
			recordCommand("READ", args)
		}
	} else {
		fmt.Println("key not found: ", key)
	}
}

func writeToStore(args ...string) {
	if len(args) < 2 {
		fmt.Println("err: WRITE command expecting two arguments - <key> <value>")
		return
	}

	key := args[0]
	value := args[1]

	if root.activeTransaction != nil {
		recordCommand("WRITE", args)
	}

	root.store[key] = value
}

func deleteFromStore(args ...string) {
	if len(args) < 1 {
		fmt.Println("err: DELETE command expecting one argument - <key>")
		return
	}

	if root.activeTransaction != nil {
		recordCommand("DELETE", args)
	}
	delete(root.store, args[0])
}

func startTransaction(args ...string) {
	newTransaction := &transaction{
		parent:   nil,
		commands: []commandLog{},
	}

	if root.activeTransaction != nil {
		newTransaction.parent = root.activeTransaction
	}

	root.activeTransaction = newTransaction
}

func commitTransaction(args ...string) {
	if root.activeTransaction == nil {
		fmt.Println("err: no active transaction found to commit")
		return
	}

	if root.activeTransaction.parent != nil {
		root.activeTransaction.parent.commands = append(root.activeTransaction.parent.commands, root.activeTransaction.commands...)
	} else {
		root.transactions = append(root.transactions, *root.activeTransaction)
	}

	root.activeTransaction = root.activeTransaction.parent
}

func abortTransaction(args ...string) {
	if root.activeTransaction != nil {
		root.activeTransaction = root.activeTransaction.parent
	} else {
		fmt.Println("err: no active transaction found to abort")
	}
}

func printTransactions(args ...string) {
	for index, transaction := range root.transactions {
		fmt.Printf("transaction #%d commands: %v\n", index+1, transaction.commands)
	}
}

func help(args ...string) {
	fmt.Println("  Avaliable commands: ")
	fmt.Println("    HELP               - prints help ")
	fmt.Println("    CLEAR              - clears the terminal ")
	fmt.Println("    READ <key>         - prints the value that stored by the key ")
	fmt.Println("    WRITE <key> <val>  - stores val with the given key ")
	fmt.Println("    DELETE <key>       - removes value that stored by the key ")
	fmt.Println("    START              - starts a transaction ")
	fmt.Println("    COMMIT             - commits a transaction ")
	fmt.Println("    ABORT              - aborts a transaction ")
	fmt.Println("    LIST               - list all commited transactions")
	fmt.Println("    QUIT               - quits from the program ")
	fmt.Println("  (commands are case-insensitive)")
	fmt.Println("Built by Besim Gurbuz")
}

func clear(args ...string) {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func main() {
	commands := map[string]interface{}{
		"READ":   readFromStore,
		"read":   readFromStore,
		"WRITE":  writeToStore,
		"write":  writeToStore,
		"DELETE": deleteFromStore,
		"delete": deleteFromStore,
		"START":  startTransaction,
		"start":  startTransaction,
		"COMMIT": commitTransaction,
		"commit": commitTransaction,
		"ABORT":  abortTransaction,
		"abort":  abortTransaction,
		"LIST":   printTransactions,
		"list":   printTransactions,
		"HELP":   help,
		"help":   help,
		"CLEAR":  clear,
		"clear":  clear,
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to K/V REPL! ")
	help()
	command := readCommand(reader)

	for ; shouldContinue(command); command = readCommand(reader) {
		commandWithArgs := strings.Split(command, " ")
		if value, exists := commands[commandWithArgs[0]]; exists {
			value.(func(...string))(commandWithArgs[1:]...)
		} else {
			printUnknownCommand(command)
		}
	}
}
