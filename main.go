package main

func main() {
	// Call the Do() method on the Check instance and retrieve any errors.
	err := check.Do()

	// If an error occurs during the execution of the Do() method,
	// log the error using the Logger stored in the Check instance and terminate the program.
	if err != nil {
		check.Logger.Fatal("Error:", err)
	}
}
