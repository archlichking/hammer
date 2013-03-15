package main 




// main func
func main() {

	NCPU := runtime.NumCPU()
	log.Println("# of CPU is ", NCPU)

	runtime.GOMAXPROCS(NCPU + 3)

	flag.Parse()
	log.Println("RPS is", initRPS)
	log.Println("Requests Limit is", requestLimit)
	log.Println("Slowness cap is", slownessLimit, "ms")

	profile = trafficprofiles.New("worldServerProfile")

	if profileFile != "" {
		profile.InitFromFile(profileFile)
	} else {
		profile.InitFromCode()
	}

	rand.Seed(time.Now().UnixNano())
}