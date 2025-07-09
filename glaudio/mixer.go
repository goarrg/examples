//go:build !goarrg_disable_gl
// +build !goarrg_disable_gl

/*
Copyright 2020 The goARRG Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"sync"
	"time"

	"goarrg.com"
	"goarrg.com/asset/audio"
)

type sfx struct {
	cursor int
	sample audio.Asset
}

type audioMixer struct {
	spec           audio.Spec
	musicFile      string
	music          audio.Asset
	musicCursor    int
	sfxs           []sfx
	bufferLength   int
	masterTrack    audio.Track
	mtx            sync.Mutex
	bufferSamples  int
	lastTime       time.Time
	pendingSamples int
}

var Mixer = &audioMixer{}

func Setup(music string) error {
	Mixer.musicFile = music
	return nil
}

func (a *audioMixer) AudioConfig() goarrg.AudioConfig {
	return goarrg.AudioConfig{
		Spec: audio.Spec{
			Channels:  audio.ChannelsStereo(),
			Frequency: 44100,
		},
	}
}

func (a *audioMixer) Init(_ goarrg.PlatformInterface, cfg goarrg.AudioConfig) error {
	a.spec = cfg.Spec

	s, err := audio.Load(Mixer.musicFile)
	if err != nil {
		return err
	}
	a.music = s
	a.masterTrack = make(audio.Track)
	a.bufferSamples = cfg.Spec.Frequency / 10

	for _, c := range a.spec.Channels {
		a.masterTrack[c] = make([]float32, a.bufferSamples)
	}

	a.lastTime = time.Now()

	return nil
}

func (a *audioMixer) Mix() (int, audio.Track) {
	Mixer.mtx.Lock()

	delta := time.Since(a.lastTime).Seconds()
	a.lastTime = time.Now()

	deltaSamples := int(delta * float64(a.spec.Frequency))
	a.pendingSamples -= deltaSamples

	if a.pendingSamples < 0 {
		a.pendingSamples = 0
	}

	samples := a.bufferSamples - a.pendingSamples
	a.pendingSamples += samples

	for i := 0; i < samples; i++ {
		cursor := (a.musicCursor + i) % a.music.DurationSamples()
		for _, c := range a.spec.Channels {
			sample := a.music.Track()[c][cursor]
			for _, sfx := range a.sfxs {
				sample += sfx.sample.Track()[c][sfx.cursor]
			}

			if sample > 1 {
				sample = 1
			}

			if sample < -1 {
				sample = -1
			}

			a.masterTrack[c][i] = sample
		}

		for i := 0; i < len(a.sfxs); {
			a.sfxs[i].cursor++

			if a.sfxs[i].cursor >= a.sfxs[i].sample.DurationSamples() {
				a.sfxs = append(a.sfxs[:i], a.sfxs[i+1:]...)
			} else {
				i++
			}
		}
	}

	if samples > 0 {
		a.musicCursor = (a.musicCursor + samples) % a.music.DurationSamples()
	}

	Mixer.mtx.Unlock()

	return samples, a.masterTrack
}

func (a *audioMixer) Update() {
}

func (a *audioMixer) Destroy() {
}

func PlaySound(sound string) error {
	s, err := audio.Load(sound)
	if err != nil {
		return err
	}

	Mixer.mtx.Lock()
	Mixer.sfxs = append(Mixer.sfxs, sfx{sample: s})
	Mixer.mtx.Unlock()

	return nil
}
