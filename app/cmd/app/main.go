package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/game"
	"github.com/vibe-code-challenge/twenty-fifty/internal/save"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	masterSeed, err := save.NewMasterSeed()
	if err != nil {
		log.Fatalf("failed to generate master seed: %v", err)
	}

	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle(game.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	saveDir := save.AutoSaveDir("")
	if err := save.EnsureSaveDir(saveDir); err != nil {
		log.Printf("warning: could not create save directory: %v", err)
	}
	savePath := save.AutoSavePath("")

	if err := ebiten.RunGame(game.New(cfg, masterSeed, savePath)); err != nil {
		log.Fatal(err)
	}
}
