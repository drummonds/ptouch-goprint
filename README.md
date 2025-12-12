# About:

This is a port of https://github.com/HenrikBengtsson/brother-ptouch-label-printer-on-linux
which was written in C and is now converted to go and support extended to E550W

It is not yet pure go but I hope it will get there eventually.  At the moment is very
experimental but I need to to work well enough to print my XMAS card labels.

ptouch-goprint is a command line tool to print labels on Brother P-Touch
printers on Linux.

There is no need to install the printer via CUPS, the printer is accessed
directly via libusb.

The tool was written for and tested with the PT-2430PC, but meanwhile is also
used with others (see "ptouch-goprint --list-supported")
Maybe others work too (please report USB VID and PID so I can include support
for further models, too).

Further info can be found at:
https://dominic.familie-radermacher.ch/projekte/ptouch-print/


## PT-550W summary

The protocol was summarised from this document [cv_pte550wp750wp710bt_eng_raster_102 downloaded from Brother](https://download.brother.com/welcome/docp100064/cv_pte550wp750wp710bt_eng_raster_102.pdf) 
and a summary in protocol_summary.md is there.

Note:

Dear visitor, currently I have absolutely no time for improvements on this
project (my free time currently is about one or two hours PER MONTH).
Therefore, I can not look at suggestions about improvements.

# Requires task from Taskfile.dev to build