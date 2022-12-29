# Introduction
Forum discussions indicate that all blue screen errors on the T120 / T520 are a result of corrupted firmware (https://h30434.www3.hp.com/t5/DesignJet-Large-Format-Printers-Digital-Press/DesignJet-T520-blue-screen-error/td-p/5879930). To replace the firmware, without purchasing a new firmware flash, please see 
https://lars.karlslund.dk/2019/01/fix-bsod-on-designjet-t120-t520-for-free/.

Not all flash drives seem to work with the printer. To ensure your flash drive will work, the first thing to do is to clone the old flash drive onto the new flash drive, and check that your printer boots and has exactly the same error message.

Note that somehow the printer *will* get new corrupt firmware despite your best effort to prevent updates. I recommend using a extension usb cable, and then placing your flash drive outside of the printer, so you do not need to unscrew the printer to access the drive again, when it becomes corrupted again. I have also tried to set my drive to read only mode (using a little toggle on the usb stick), but that prevents the printer from booting properly.

# Shortcut (cloning the pre-saved image to a usb drive)
**Warning** It is critical that you never assume the location of the flash drives (/dev/disk3), and always check before doing any dd commands. dd used with an incorrect link could instantly begin wiping out a partition on your computer, totally destroying it.

I have created a flash image with the files below, in a single flash image, this should be enough to fix your bsod, depending on the version of the printer.

1. plug in your flash drive
2. unmount the flash drive, first `diskutil list` to see the available drives (and associated partitions), then unmount the flash drive `diskutil unmountDisk disk3`.
3. clone the t120flash.img to the flash drive, `sudo dd if=t120flash.img of=/dev/disk3 bs=4m status=progress`
4. confirm the flash drive has the new correct structure, `diskutil list`, three partitions, 

```
/dev/disk3 (external, physical):
   #:                       TYPE NAME                    SIZE       IDENTIFIER
   0:     FDisk_partition_scheme                        *2.0 GB     disk3
   1:                 DOS_FAT_32                         33.6 MB    disk3s1
   2:                 DOS_FAT_32                         33.6 MB    disk3s2
   3:                 DOS_FAT_32 NO NAME                 1.9 GB     disk3s3
```

# Cloning the usb drive (partition structure)
Mount both the original and new flash drive, and make sure that both are visible using `diskutil list`. Note that the 'new' flash is 4GB, some old scrap drive I found lying around.
```
diskutil list
/dev/disk2 (external, physical):
   #:                       TYPE NAME                    SIZE       IDENTIFIER
   0:     FDisk_partition_scheme                        *4.1 GB     disk2
   1:                 DOS_FAT_32 NO NAME                 4.1 GB     disk2s1

/dev/disk3 (external, physical):
   #:                       TYPE NAME                    SIZE       IDENTIFIER
   0:     FDisk_partition_scheme                        *2.0 GB     disk3
   1:                 DOS_FAT_32                         33.6 MB    disk3s1
   2:                 DOS_FAT_32                         33.6 MB    disk3s2
   3:                 DOS_FAT_32 NO NAME                 1.9 GB     disk3s3
```
use the dd command to copy /dev/disk3 to /dev/disk2. This is important to create the same partition structure on the new drive. 
**Note** HP seems to have some magic with the partitions, and I was never able to create a working flash drive from a new usb, despite my fdisk structure appearing to be exactly the same as the old flash drive.
**Warning** It is critical that you never assume the location of the flash drives (/dev/disk2), and always check before doing any dd commands. dd used with an incorrect link could instantly begin wiping out a partition on your computer, totally destroying it.
```
$ sudo dd if=/dev/disk3 of=/dev/disk2 bs=4m status=progress
477+1 records in
477+1 records out
2003828736 bytes transferred in 866.017384 secs (2313844 bytes/sec)
```
# Testing the new flash drive
Put the flash into the printer and make sure that you get the same error message. A flash drive that is not compatible will cause the printer to emit a beep sound. Once you have the same error message on the new flash drive, you can now keep the old flash drive safe, as a backup.

# Creating the new firmware image (optional)
Note that I have provided an already fixed file in this repo (AXP2CN1829BR.bin), so depending on your printer, you might not need to do this step at all, you can skip to the `Write the firmware image to the new flash` section.

1. Save the go code to a hpfix.go file, and modify the target file name to the firmware that corresponding to the name as on the original flash eg. 'AXP2CN1829BR.bin'
2. You can find amperexl_pr_AXP2CN2022AR_secure_signed_rbx.ful here https://mnogochernil.ru/newsroom/hp-designjet-t120-t520-firmware-versions/ (added to repo as its not on the RU site anymore)
3. Run the go file, `go run hpfix.go`, you might need to import the library, in which case you would also need `go get -u github.com/davecgh/go-spew/spew`
4. inspect the resulting file, and make sure it has a valid header (feed f00d) as described in https://lars.karlslund.dk/2019/01/fix-bsod-on-designjet-t120-t520-for-free/
```
$ xxd AXP2CN1829BR.bin|head                                                                                                                                                                                                                     2 ↵
00000000: feed f00d 8350 0000 01af fff8 84ff d760  .....P.........`
00000010: 04e0 2de5 5c10 9fe5 0010 91e5 0000 a0e3  ..-.\...........
00000020: b000 d1e1 04e0 9de4 1eff 2fe1 0300 a0e3  ........../.....
00000030: 1200 00ea 7847 46c0 1040 2de9 0040 a0e1  ....xGF..@-..@..
00000040: f2ff ffeb 0400 a0e1 1040 bde8 0b00 00ea  .........@......
00000050: 7847 46c0 0600 a0e3 0800 00ea 7847 46c0  xGF.........xGF.
00000060: 1040 2de9 0040 a0e1 e8ff ffeb 0400 a0e1  .@-..@..........
00000070: 1040 bde8 0100 00ea 8000 0004 7847 46c0  .@..........xGF.
00000080: 0010 0fe1 c020 81e3 02f0 21e1 2020 9fe5  ..... ....!.  ..
00000090: 0020 92e5 0030 a0e3 b030 d2e1 b000 c2e1  . ...0...0......
```

# Write the firmware image to the new flash
Check the new flash is properly mounted
```
$ diskutil list
/dev/disk2 (external, physical):
   #:                       TYPE NAME                    SIZE       IDENTIFIER
   0:     FDisk_partition_scheme                        *4.1 GB     disk2
   1:                 DOS_FAT_32                         33.6 MB    disk2s1
   2:                 DOS_FAT_32                         33.6 MB    disk2s2
   3:                 DOS_FAT_32 NO NAME                 1.9 GB     disk2s3
```
Then unmount the disk
```
$ diskutil unmountDisk /dev/disk2
Unmount of all volumes on disk2 was successful
```

:warning: once again I remind you to always triple check the disk you intend to dd, selecting the wrong disk will destroy it.

Then dd the image to *both* flash partitions
```
$ sudo dd if=AXP2CN1829BR.bin of=/dev/disk2s1 bs=4m status=progress
Password:
6+1 records in
6+1 records out
28327670 bytes transferred in 12.911111 secs (2194054 bytes/sec)
```
and 
```
$ sudo dd if=AXP2CN1829BR.bin of=/dev/disk2s2 bs=4m status=progress
6+1 records in
6+1 records out
28327670 bytes transferred in 12.995905 secs (2179738 bytes/sec)
```
Note that on an old flash drive can take over 2 minutes (also leaving out the bs=4m makes it take _much_ longer!)

# Testing the new firmware
Put the flash into the printer and you should be able to boot past the error message, and your printer will be working again.

# Resources
```
28327670 Jan  3 14:44 AXP2CN1829BR.bin
    4985 Jan  3 14:42 README.md
25951768 Jan  3 14:44 amperexl_plus_pr_APP2CN1829BR_secure_signed_rbx.ful
25941028 Jan  3 14:43 amperexl_pr_AXP2CN1829BR_secure_signed_rbx.ful
    6962 Jan  3 14:42 hpfix.go
26018267 Jan  3 14:47 t120_fw_backup.tgz
```
`AXP2CN1829BR.bin` the result of running the go code for `amperexl_pr_AXP2CN1829BR_secure_signed_rbx.ful`
`hpfix.go` the go code from https://lars.karlslund.dk/2019/01/fix-bsod-on-designjet-t120-t520-for-free/
`t120_fw_backup.tgz` a backup of the original flash drive with the firmware error.

