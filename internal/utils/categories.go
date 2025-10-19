package utils

import "strings"

// ProgramCategories maps category names to program keywords
var ProgramCategories = map[string][]string{
	"IT": {
		"Artificial Intelligence", "Cybersecurity", "Digital Infrastructure", "ICT",
		"Software Development", "Information Systems",
		"Web Development", "Mobile App", "Computer Science", "Data Science",
		"Information Technology", "Software Engineering",
		"Information and Communications Technology",
		"AI", "ITM", "IT", "CS",
	},
	"Engineering": {
		"Civil Engineering", "Mechanical Engineering", "Electrical Engineering",
		"Chemical Engineering", "Aerospace Engineering", "Industrial Engineering",
		"Engineering",
	},
	"Business": {
		"Business Administration", "MBA", "Marketing", "Finance",
		"Accounting", "Management", "Business", "Commerce",
	},
	"Medicine": {
		"Medicine", "Nursing", "Pharmacy", "Dentistry",
		"Public Health", "Medical", "Healthcare",
	},
	"Science": {
		"Biology", "Chemistry", "Physics", "Mathematics",
		"Environmental Science", "Biotechnology", "Science",
	},
	"Arts": {
		"Fine Arts", "Design", "Music", "Theater",
		"Architecture", "Graphic Design", "Art",
	},
	"Economics": {
		"Economics", "Economic", "Finance",
	},
}

// ExpandCategoryToPrograms returns the list of program keywords
// associated with a given category name (case-insensitive)
func ExpandCategoryToPrograms(category string) []string {
	category = strings.TrimSpace(strings.ToLower(category))

	for key, programs := range ProgramCategories {
		if strings.ToLower(key) == category {
			return programs
		}
	}

	return []string{}
}

// ValidCategories returns a slice of all available category names
func ValidCategories() []string {
	categories := make([]string, 0, len(ProgramCategories))
	for category := range ProgramCategories {
		categories = append(categories, category)
	}
	return categories
}

