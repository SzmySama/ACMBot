package db

import "gorm.io/gorm"

type CodeforcesUser struct {
	gorm.Model
}

type CodeforcesSubmission struct {
	gorm.Model
}

type CodeforcesProblem struct {
	gorm.Model
}

type CodeforcesRatingChange struct {
	gorm.Model
}
