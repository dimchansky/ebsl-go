package solver

import (
	"errors"
	"fmt"
	"math"

	"github.com/dimchansky/ebsl-go/opinion"
	"github.com/dimchansky/ebsl-go/trust/equations"
)

var (
	ErrEpochMustBePositiveNumber = errors.New("solver: epoch must be positive number")
)

type distanceAggregator interface {
	Reset()
	Add(float64)
	Result() float64
}

type options struct {
	epochs             uint
	distanceFun        func(prevValue *opinion.Type, newValue *opinion.Type) float64
	distanceAggregator distanceAggregator
	tolerance          float64
}

type Options func(opts *options) (*options, error)

func Epochs(epochs uint) Options {
	return func(opts *options) (*options, error) {
		if epochs < 1 {
			return nil, ErrEpochMustBePositiveNumber
		}
		opts.epochs = epochs
		return opts, nil
	}
}

func ManhattanDistance() Options {
	return func(opts *options) (*options, error) {
		opts.distanceFun = manhattanDistance
		return opts, nil
	}
}

func ChebyshevDistance() Options {
	return func(opts *options) (*options, error) {
		opts.distanceFun = chebyshevDistance
		return opts, nil
	}
}

func EuclideanDistance() Options {
	return func(opts *options) (*options, error) {
		opts.distanceFun = euclideanDistance
		return opts, nil
	}
}

func Tolerance(tolerance float64) Options {
	return func(opts *options) (*options, error) {
		opts.tolerance = tolerance
		return opts, nil
	}
}

func SolveEquations(context equations.Context, eqs equations.Equations, opts ...Options) (err error) {
	solverOpts := &options{
		epochs:             100,
		distanceFun:        manhattanDistance,
		distanceAggregator: &maxDistanceAggregator{},
		tolerance:          0.0,
	}

	// apply solver options
	for _, applyOption := range opts {
		solverOpts, err = applyOption(solverOpts)
		if err != nil {
			return
		}
	}

	epochs := solverOpts.epochs
	distanceFun := solverOpts.distanceFun
	distanceAggregator := solverOpts.distanceAggregator
	tolerance := solverOpts.tolerance

	for epoch := uint(1); epoch <= epochs; epoch++ {
		fmt.Printf("Epoch: %v\n", epoch) // TODO: add callback

		distanceAggregator.Reset()
		for _, eq := range eqs {
			prevValue := *context.GetFinalReferralTrust(eq.R)
			newValue, err := eq.Evaluate(context)
			if err != nil {
				return err
			}
			dist := distanceFun(&prevValue, newValue)

			// fmt.Printf("R[%v,%v]: prev: %v new: %v dist: %v\n", eq.R.From, eq.R.To, prevValue, newValue, dist) // TODO: add callback

			distanceAggregator.Add(dist)
		}
		distError := distanceAggregator.Result()
		fmt.Printf("Epoch %v error: %v\n", epoch, distError) // TODO: add callback

		if distError <= tolerance {
			return nil
		}
	}

	return nil
}

func manhattanDistance(prevValue *opinion.Type, newValue *opinion.Type) float64 {
	return math.Abs(prevValue.B-newValue.B) +
		math.Abs(prevValue.D-newValue.D) +
		math.Abs(prevValue.U-newValue.U)
}

func chebyshevDistance(prevValue *opinion.Type, newValue *opinion.Type) float64 {
	return math.Max(
		math.Max(
			math.Abs(prevValue.B-newValue.B),
			math.Abs(prevValue.D-newValue.D),
		),
		math.Abs(prevValue.U-newValue.U),
	)
}

func euclideanDistance(prevValue *opinion.Type, newValue *opinion.Type) float64 {
	return math.Sqrt(square(prevValue.B-newValue.B) +
		square(prevValue.D-newValue.D) +
		square(prevValue.U-newValue.U),
	)
}

func square(x float64) float64 { return x * x }

type maxDistanceAggregator struct {
	nonEmpty    bool
	maxDistance float64
}

func (a *maxDistanceAggregator) Reset() {
	a.nonEmpty = false
	a.maxDistance = math.MaxFloat64
}

func (a *maxDistanceAggregator) Add(v float64) {
	if a.nonEmpty {
		a.maxDistance = math.Max(a.maxDistance, v)
	} else {
		a.nonEmpty = true
		a.maxDistance = v
	}
}

func (a *maxDistanceAggregator) Result() float64 {
	return a.maxDistance
}
