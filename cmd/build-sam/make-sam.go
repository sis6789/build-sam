package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	. "github.com/sis6789/global_definition/align"
	"github.com/sis6789/global_definition/human_reader"
	"github.com/sis6789/global_definition/localip"
	"github.com/sis6789/global_definition/machine_data"
	"github.com/sis6789/simple_mongo/keydb2"
	"go.mongodb.org/mongo-driver/bson"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/sis6789/build-sam/internal/app/build-sam"
)

// 운영 환경 설정
var flagDb = flag.StringP("db", "d", "", "-d dbname")
var flagGenome = flag.BoolP("genome", "g", false, "-g")
var flagMolecular = flag.BoolP("molecular", "m", false, "-m")
var flagHost = flag.StringP("server", "s", "192.168.0.6", "-s address")
var flagOutput = flag.StringP("output", "o", "", "-o path")

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	var err error

	// 명령줄 flag를 설정한다.
	flag.Parse()
	err = viper.BindPFlags(flag.CommandLine)
	if err != nil {
		log.Printf("%v", err)
	}
	var outputFolder = viper.GetString("output")
	outputFolder, _ = homedir.Expand(outputFolder)
	outputFolder, _ = filepath.Abs(outputFolder)
	err = os.MkdirAll(outputFolder, 0777)
	if err != nil {
		log.Fatalf("%v", err)
	}
	var dbName = viper.GetString("db")
	if dbName == "" {
		log.Fatalf("specify -d parameter")
	}
	var colName string
	var getQueryName build_sam.QueryName
	var logFN string
	switch {
	case viper.GetBool("molecular"):
		colName = "molecular"
		getQueryName = func(r FmdsRaw) string {
			return fmt.Sprintf("%v-%v-%v", r.Name, r.FileName, r.Molecular)
		}
		logFN = viper.GetString("db") + "-" + "molecular.log"
	case viper.GetBool("genome"):
		colName = "genome"
		getQueryName = func(r FmdsRaw) string {
			return r.Name
		}
		logFN = viper.GetString("db") + "-" + "genome.log"
	default:
		log.Fatalf("%v", "specify one of -g -m flag")
	}
	var samPath = filepath.Join(outputFolder, dbName+"-"+colName+".sam")
	var bamPath = filepath.Join(outputFolder, dbName+"-"+colName+".bam")

	logPath := filepath.Join(outputFolder, logFN)
	if err != nil {
		log.Fatalf("%v", err)
	}
	logFD, err := os.Create(logPath)
	if err != nil {
		log.Fatalf("%v", err)
	}
	mw := io.MultiWriter(logFD, os.Stdout)
	log.SetOutput(mw)
	log.Print("************************************")
	log.Print("SAM/BAM output from Align Result")
	log.Print("Copyright 2022. KEYOMICS inc. all rights reserved.")
	log.Print("------------------------------------")
	absCommand, _ := filepath.Abs(os.Args[0])
	if len(os.Args) >= 2 {
		log.Printf("%v %v", absCommand, strings.Join(os.Args[1:], " "))
	} else {
		log.Printf("%v", absCommand)
	}
	log.Print("------------------------------------")

	var mongoURI = fmt.Sprintf("mongodb://%v:27017", viper.GetString("server"))

	startJob := time.Now()

	var client = keydb2.New(mongoURI)
	var colSource = client.Col(dbName, colName)
	var ctx = context.TODO()

	hostIp, err := localip.LocalIP()
	if err != nil {
		log.Printf("%v", err)
	}
	humanFolder, err := machine_data.MachineData(hostIp, "humanPath")
	if err != nil {
		log.Printf("%v", err)
	}
	human_reader.Prepare(humanFolder)

	// output file
	bioSamFile, err := os.Create(samPath)
	if err != nil {
		log.Fatalf("%v", err)
	}
	bioSam := build_sam.PrepareSam(bioSamFile)
	bioBamFile, err := os.Create(bamPath)
	if err != nil {
		log.Fatalf("%v", err)
	}
	bioBam := build_sam.PrepareBam(bioBamFile)

	// read into memory
	log.Printf("start db read")
	start := time.Now()
	cursor, err := colSource.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("%v", err)
	}
	build_sam.StartConverter()
	readCount := 0
	for cursor.Next(ctx) {
		var thisRec FmdsRaw
		err = cursor.Decode(&thisRec)
		build_sam.ChanFmds <- thisRec
		readCount++
		if readCount%10_0000 == 0 {
			log.Printf("db read %d", readCount)
		}
	}
	// close cursor
	err = cursor.Close(ctx)
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("db read %v for %v", readCount, time.Since(start))
	close(build_sam.ChanFmds)
	build_sam.WgChanFmds.Wait()
	close(build_sam.ChanMB)
	build_sam.WgChanMB.Wait()
	err = cursor.Close(ctx)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// save each record
	build_sam.SaveSam(bioSam, bioBam, getQueryName)

	// close
	err = bioBam.Close()
	if err != nil {
		log.Fatalf("%v", err)
	}
	for _, v := range []*os.File{bioSamFile, bioBamFile} {
		err = v.Close()
		if err != nil {
			log.Fatalf("%v", err)
		}
	}

	human_reader.Close()

	start = time.Now()
	log.Printf("start bam index")

	bioBamIndexFile, err := os.Open(bamPath)
	if err != nil {
		log.Fatalf("%v", err)
	}
	build_sam.MakeBamIndex(bioBamIndexFile, bamPath)
	err = bioBamIndexFile.Close()
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("bam index for %v", time.Since(start))

	log.Printf("EOJ for %v", time.Since(startJob))
}

func unused() {

}
