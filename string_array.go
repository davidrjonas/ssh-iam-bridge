package main

func stringArrayContains(test string, set []string) bool {
	for _, member := range set {
		if test == member {
			return true
		}
	}

	return false
}

// Compares array1 against a2 and returns the values in array1 that are not present in a2.
func stringArrayDiff(tests, set []string) []string {

	missing := make([]string, 0)

	for _, test := range tests {
		if stringArrayContains(test, set) {
			continue
		}

		missing = append(missing, test)
	}

	return missing
}
