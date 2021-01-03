package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

type stateinfo uint8

const (
	LookingForESC stateinfo = iota
	LookingForStar
	LookingForUpperCase
	Data
	DataSkip
)

type decodemode uint8

const (
	Normal decodemode = iota // b
	RLE                      // b2m
	Zeroes
)

func main() {
	debug := false
	f, err := os.Open("amperexl_pr_AXP2CN2022AR_secure_signed_rbx.ful")
	if err != nil {
		fmt.Printf("Problem opening source file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	s := bufio.NewReader(f)

	tf, err := os.Create("AXP2CN2022AR.bin")
	if err != nil {
		fmt.Printf("Problem creating destination file: %v\n", err)
		os.Exit(1)
	}
	defer tf.Close()
	t := bufio.NewWriter(tf)
	defer t.Flush()

	pos := 0
	var blocksize, totalsize int
	var state stateinfo
	var dm decodemode
	var datasize, dataleft int
	var firstfound, problem bool
	var commandpos int
	var command string
	// var commands int
	output := make([]byte, 0, 16384)

	var problems, blocks0, blocks1, blocks2, blocks3, blocksp, bytesread int

	fmt.Println("Processing ...")

	var skipfirst int
	if !debug {
		// Header
		t.Write([]byte{0xFE, 0xED, 0xF0, 0x0D})
		skipfirst = 0x10E // This might need correction
	}

	//	var header = make([]byte, 5)
	//	var expectedheader = []byte{0xFE, 0xED, 0xF0, 0x0A, 0x02}

	// s.Read(header)
	// if !bytes.Equal(header, expectedheader) {
	// 	fmt.Printf("Wrong header detected - I got %0X, but expected %0X\n", header, expectedheader)
	// 	os.Exit(1)
	// }

mainloop:
	for {
		switch state {
		case LookingForESC:
			d, err := s.ReadByte()
			if err != nil {
				fmt.Printf("Problem reading data: %v\n", err)
				break mainloop
			}
			if d == 0x1B {
				problem = false
				command = ""
				state++
			} else if firstfound && !problem {
				fmt.Print("Not at ESC\n")
				problem = true
				problems++
			}
		case LookingForStar:
			d, err := s.ReadByte()
			if err != nil {
				fmt.Printf("Problem reading data: %v\n", err)
				break mainloop
			}
			if d == 0x2A {
				commandpos = pos - 1
				state++
				firstfound = true
			} else {
				state = LookingForESC
			}
		case LookingForUpperCase:
			d, err := s.ReadByte()
			if err != nil {
				fmt.Printf("Problem reading data: %v\n", err)
				break mainloop
			}
			command = command + string(d)
			if d >= 'A' && d <= 'Z' {
				state = LookingForESC
				switch {
				case strings.HasPrefix(command, "rt"):
					blocksize, _, err = getint(command[2:])
					if debug {
						fmt.Fprintf(t, "%08X: *%v - blocksize set to %v\n", commandpos, command, blocksize)
					}
				case command == "rC":
					if debug {
						fmt.Fprintf(t, "%08X: *%v - end of update\n", commandpos, command)
					}
					break mainloop
				case strings.HasPrefix(command, "b+1ym") || strings.HasPrefix(command, "b+2ym"):
					dm = Normal
					if debug {
						fmt.Fprintf(t, "%08X: *%v - checksum\n", commandpos, command)
					}
					datasize, command, err = getint(command[5:])
					if err != nil {
						fmt.Printf("Error parsing command, integer not found: %v\n", command)
						break mainloop
					}
					dataleft = datasize
					state = DataSkip
				case strings.HasPrefix(command, "b+"):
					dm = Normal
					if debug {
						fmt.Fprintf(t, "%08X: *%v - block+\n", commandpos, command)
					}
					blocksp++
					datasize, command, err = getint(command[2:])
					if err != nil {
						fmt.Printf("Error parsing command, integer not found: %v\n", command)
						break mainloop
					}
					if datasize > 0 {
						dataleft = datasize
						state = Data
					}
				case strings.HasPrefix(command, "b"):
					rest := command[1:]
					switch {
					case strings.HasPrefix(rest, "0m"):
						fmt.Println("Switching to normal")
						dm = Normal
						rest = rest[2:]
					case strings.HasPrefix(rest, "2m"):
						fmt.Println("Switching to RLE")
						dm = RLE
						rest = rest[2:]
					case strings.HasPrefix(rest, "3m"):
						fmt.Println("Switching to Zeroes")
						dm = Zeroes
						// dm = Normal
						rest = rest[2:]
					}
					if debug {
						fmt.Fprintf(t, "%08X: *%v - block\n", commandpos, command)
					}
					blocks1++
					datasize, rest, err = getint(rest)
					if err != nil {
						fmt.Printf("Error parsing command, integer not found: %v\n", command)
						break mainloop
					}
					dataleft = datasize
					state = Data
				default:
					fmt.Printf("%08X: *%v - UNKNOWN\n", commandpos, command)
					os.Exit(1)
				}
			}
		case DataSkip:
			if dataleft != 0 {
				_, err := s.ReadByte()
				if err != nil {
					fmt.Printf("Problem reading data: %v\n", err)
					break mainloop
				}
				dataleft--
			}
			if dataleft == 0 {
				state = LookingForESC
			}
		case Data:
			if dataleft > 0 {
				d, err := s.ReadByte()
				if err != nil {
					fmt.Printf("Problem reading data: %v\n", err)
					break mainloop
				}
				bytesread++
				dataleft--
				switch dm {
				case Zeroes:
					// First three are normal chars
					// Next five are zeroes repeated
					normals := int(d>>5) + 1
					zeroes := int(d & 0x1F)
					if zeroes == 0x1F {
						for {
							morezeroes, _ := s.ReadByte()
							zeroes += int(morezeroes)
							dataleft--
							if morezeroes != 0xFF {
								break
							}
						}
					}
					for i := 0; i < zeroes; i++ {
						output = append(output, 0)
					}
					for i := 0; i < normals; i++ {
						d, _ = s.ReadByte()
						dataleft--
						output = append(output, d)
					}
				case Normal:
					output = append(output, d)
				case RLE:
					if d > 0x80 {
						// Repeat next character 0xFF - d + 2 times
						repeat := 0xFF - int(d) + 2
						d, _ = s.ReadByte()
						dataleft--
						for i := 0; i < repeat; i++ {
							output = append(output, d)
						}
					} else {
						// Next d+1 chars are regular
						for i := 0; i <= int(d); i++ {
							n, _ := s.ReadByte()
							dataleft--
							output = append(output, n)
						}
					}
				}
			}
			if dataleft == 0 {
				state = LookingForESC
				if len(output) != 0 && len(output) != blocksize {
					fmt.Printf("Incorrect block size detected: %v\n", len(output))
				}
				if debug {
					spew.Fdump(t, output[:blocksize])
				} else {
					t.Write(output[skipfirst:blocksize])
					skipfirst = 0
				}
				totalsize += blocksize
				output = make([]byte, 0, 16384)
			}
		}
		if state == Data && strings.HasSuffix(command, "Y") {
			state = LookingForESC
			dataleft = 0
			datasize = 0
		}
		pos++
	}
	fmt.Printf("Problems: %v - Blocks: %v %v %v %v %v - total %v blocks - %v bytes\n", problems, blocks0, blocks1, blocks2, blocks3, blocksp, blocks0+blocks1+blocks2+blocks3+blocksp, bytesread)
	fmt.Printf("Total size: %v\n", totalsize)
}

var NoIntegerFound = errors.New("No integer found")

func getint(s string) (num int, rest string, err error) {
	var is string
	for len(s) > 0 && s[0] >= '0' && s[0] <= '9' {
		is = is + string(s[0])
		s = s[1:]
	}
	if is == "" {
		err = NoIntegerFound
		return
	}
	num, err = strconv.Atoi(is)
	rest = s
	return
}