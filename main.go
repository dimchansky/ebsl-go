package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/dimchansky/ebsl-go/evidence"
	"github.com/dimchansky/ebsl-go/trust"
	"github.com/dimchansky/ebsl-go/trust/equations"
	"github.com/dimchansky/ebsl-go/trust/equations/solver"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("Usage: %s <threshold> <evidence_file_name> <final_referral_trust_output_file>\n", os.Args[0])
		os.Exit(1)
	}

	threshold, inputFileName, outputFileName := parseCmdLineParams()

	parser := evidenceFileParser{inputFileName}

	dro := make(trust.DirectReferralOpinion).FromIterableEvidences(parser, threshold)

	log.Println("Creating Final Referral Trust equations...")
	eqs := equations.CreateFinalReferralTrustEquations(dro)
	log.Println("Final Referral Trust equations are created.")

	context := equations.NewDefaultFinalReferralTrustEquationContext(dro)

	log.Println("Solving Final Referral Trust equations...")
	if err := solver.SolveFinalReferralTrustEquations(
		context,
		eqs,
		solver.UseOnEpochEndCallback(func(epoch uint, aggregatedDistance float64) error {
			log.Printf("Epoch %v error: %v\n", epoch, aggregatedDistance)
			return nil
		}),
	); err != nil {
		log.Fatal(err)
	}
	log.Println("Final Referral Trust equations are solved.")

	log.Println("Writing final referral trust discount values to file...")
	if err := writeFinalReferralTrustDiscount(outputFileName, context); err != nil {
		fmt.Printf("failed to write final referral trust discounts to file: %v\n", err)
		os.Exit(2)
	}
	log.Println("Done.")
}

func parseCmdLineParams() (threshold uint64, inputFileName string, outputFileName string) {
	thresholdStr := os.Args[1]
	c, err := strconv.Atoi(thresholdStr)
	if err != nil {
		fmt.Printf("invalid threshold value (%v): %v", thresholdStr, err)
		os.Exit(1)
	}
	if c <= 0 {
		fmt.Println("threshold value must be positive number")
		os.Exit(1)
	}
	threshold = uint64(c)
	inputFileName = os.Args[2]
	outputFileName = os.Args[3]
	return
}

func writeFinalReferralTrustDiscount(outputFileName string, context *equations.DefaultFinalReferralTrustEquationContext) (err error) {
	outFile, err := os.Create(outputFileName)
	if err != nil {
		return
	}
	defer func() {
		if tErr := outFile.Close(); tErr != nil && err == nil {
			err = tErr
		}
	}()
	of := bufio.NewWriter(outFile)
	defer func() {
		if tErr := of.Flush(); tErr != nil && err == nil {
			err = tErr
		}
	}()

	for key, value := range context.FinalReferralTrust {
		_, err = of.WriteString(fmt.Sprintf("%v\t%v\t%v\n", key.From, key.To, context.GetDiscount(value)))
		if err != nil {
			return
		}
	}
	return
}

type evidenceFileParser struct {
	fileName string
}

func (p evidenceFileParser) GetEvidenceIterator() trust.EvidenceIterator {
	return parseEvidenceFile(p.fileName)
}

func parseEvidenceFile(fileName string) trust.EvidenceIterator {
	return func(onNext trust.NextEvidenceHandler) (err error) {
		inputFile, err := os.Open(fileName)
		if err != nil {
			return err
		}
		defer func() {
			if tErr := inputFile.Close(); tErr != nil && err == nil {
				err = tErr
			}
		}()

		sc := bufio.NewScanner(bufio.NewReader(inputFile))
		for sc.Scan() {
			fields := bytes.Fields(sc.Bytes())
			if len(fields) != 4 {
				continue
			}

			from, err := strconv.Atoi(string(fields[0]))
			if err != nil {
				return err
			}
			if from < 0 {
				return errors.New("link source must be non-negative number")
			}
			to, err := strconv.Atoi(string(fields[1]))
			if err != nil {
				return err
			}
			if to < 0 {
				return errors.New("link destination must be non-negative number")
			}
			pos, err := strconv.ParseFloat(string(fields[2]), 64)
			if err != nil {
				return err
			}
			neg, err := strconv.ParseFloat(string(fields[3]), 64)
			if err != nil {
				return err
			}

			if err := onNext(trust.Link{From: uint64(from), To: uint64(to)}, evidence.New(pos, neg)); err != nil {
				return err
			}
		}

		return sc.Err()
	}
}
