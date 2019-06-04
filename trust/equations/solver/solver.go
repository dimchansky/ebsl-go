package solver

import (
	"errors"
	"math"
	"sort"

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

func UseManhattanDistance() Options {
	return func(opts *options) (*options, error) {
		opts.distanceFun = manhattanDistance
		return opts, nil
	}
}

func UseChebyshevDistance() Options {
	return func(opts *options) (*options, error) {
		opts.distanceFun = chebyshevDistance
		return opts, nil
	}
}

func UseEuclideanDistance() Options {
	return func(opts *options) (*options, error) {
		opts.distanceFun = euclideanDistance
		return opts, nil
	}
}

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

func SolveEquations(context equations.Context, eqs equations.Equations, opts ...Options) (err error) {
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

	// order equations by direct first, then by indices
	orderEquationsByDirectRefAndIndices(eqs)

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
		if err := onEpochEnd(epoch, distError); err != nil {
			return err
		}

		if distError <= tolerance {
			return nil
		}
	}

	return nil
}

// orderEquationsByDirectRefAndIndices orders equations so that direct referral equations go first and the all equations are ordered by indices of R
func orderEquationsByDirectRefAndIndices(rEquations equations.Equations) {
	sort.Slice(rEquations, func(i, j int) bool {
		iEq := rEquations[i]
		jEq := rEquations[j]
		iExpDirect := iEq.Expression.IsDirectReferralTrust()
		jExpDirect := jEq.Expression.IsDirectReferralTrust()
		if iExpDirect != jExpDirect {
			return iExpDirect // direct equations go first
		}

		// the sort by R indices
		iFrom := iEq.R.From
		jFrom := jEq.R.From
		if iFrom != jFrom {
			return iFrom < jFrom
		}

		return iEq.R.To < jEq.R.To
	})
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