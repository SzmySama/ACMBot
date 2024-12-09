package model

type RaceProvider func() ([]Race, error)

type PicProvider func() ([]byte, error)
