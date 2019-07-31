package solver

import (
	"errors"
	"math"

	"github.com/dimchansky/ebsl-go/opinion"
	"github.com/dimchansky/ebsl-go/trust/equations"
)

var (
	ErrEpochMustBePositiveNumber = errors.New("solver: epoch must be positive number")
)

type DistanceFun func(prevValue *opinion.Type, newValue *opinion.Type) float64

type DistanceAggregator interface {
	Reset()
	Add(float64)
	Result() float64
}

type EpochStartFun func(epoch uint) error
type EpochEndFun func(epoch uint, aggregatedDistance float64) error

type options struct {
	epochs             uint
	distanceFun        DistanceFun
	distanceAggregator DistanceAggregator
	tolerance          float64
	onEpochStart       EpochStartFun
	onEpochEnd         EpochEndFun
}

type Options func(opts *options) (*options, error)

func UseMaxEpochs(epochs uint) Options {
	return func(opts *options) (*options, error) {
		if epochs < 1 {
			return nil, ErrEpochMustBePositiveNumber
		}
		opts.epochs = epochs
		return opts, nil
	}
}

func UseManhattanDistance() Options { return UseDistanceFunction(manhattanDistance) }
func UseChebyshevDistance() Options { return UseDistanceFunction(chebyshevDistance) }
func UseEuclideanDistance() Options { return UseDistanceFunction(euclideanDistance) }

func UseMaxDistanceAggregator() Options { return UseDistanceAggregator(&maxDistanceAggregator{}) }
func UseSumDistanceAggregator() Options { return UseDistanceAggregator(&sumDistanceAggregator{}) }

func UseDistanceFunction(distanceFun DistanceFun) Options {
	return func(opts *options) (*options, error) {
		opts.distanceFun = distanceFun
		return opts, nil
	}
}

func UseDistanceAggregator(distanceAggregator DistanceAggregator) Options {
	return func(opts *options) (*options, error) {
		opts.distanceAggregator = distanceAggregator
		return opts, nil
	}
}

func UseTolerance(tolerance float64) Options {
	return func(opts *options) (*options, error) {
		opts.tolerance = tolerance
		return opts, nil
	}
}

func UseOnEpochStartCallback(onEpochStart EpochStartFun) Options {
	return func(opts *options) (*options, error) {
		opts.onEpochStart = onEpochStart
		return opts, nil
	}
}

func UseOnEpochEndCallback(onEpochEnd EpochEndFun) Options {
	return func(opts *options) (*options, error) {
		opts.onEpochEnd = onEpochEnd
		return opts, nil
	}
}

func SolveFinalReferralTrustEquations(
	context equations.FinalReferralTrustEquationContext,
	eqs equations.IterableFinalReferralTrustEquations,
	opts ...Options,
) (err error) {
	solverOpts := &options{
		epochs:             100,
		distanceFun:        manhattanDistance,
		distanceAggregator: &maxDistanceAggregator{},
		tolerance:          0.0,
		onEpochStart: func(epoch uint) error {
			return nil
		},
		onEpochEnd: func(epoch uint, aggregatedDistance float64) error {
			return nil
		},
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
	onEpochStart := solverOpts.onEpochStart
	onEpochEnd := solverOpts.onEpochEnd

	for epoch := uint(1); epoch <= epochs; epoch++ {
		if err := onEpochStart(epoch); err != nil {
			return err
		}

		distanceAggregator.Reset()

		foreachEquation := eqs.GetFinalReferralTrustEquationIterator()
		err = foreachEquation(func(eq *equations.FinalReferralTrustEquation) error {
			prevValue := *context.GetFinalReferralTrust(eq.R)
			newValue, err := eq.EvaluateFinalReferralTrust(context)
			if err != nil {
				return err
			}
			dist := distanceFun(&prevValue, newValue)

			// fmt.Printf("R[%v,%v]: prev: %v new: %v dist: %v\n", eq.R.From, eq.R.To, prevValue, newValue, dist) // TODO: add callback

			distanceAggregator.Add(dist)
			return nil
		})
		if err != nil {
			return
		}

		distError := distanceAggregator.Result()
		if err := onEpochEnd(epoch, distError); err != nil {
			return err
		}

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

type sumDistanceAggregator struct {
	nonEmpty    bool
	sumDistance float64
}

func (a *sumDistanceAggregator) Reset() {
	a.nonEmpty = false
	a.sumDistance = math.MaxFloat64
}

func (a *sumDistanceAggregator) Add(v float64) {
	if a.nonEmpty {
		a.sumDistance += v
	} else {
		a.nonEmpty = true
		a.sumDistance = v
	}
}

func (a *sumDistanceAggregator) Result() float64 {
	return a.sumDistance
}
