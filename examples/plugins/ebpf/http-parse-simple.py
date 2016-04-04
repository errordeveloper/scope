#!/usr/bin/python
#
#Bertrone Matteo - Polytechnic of Turin
#November 2015
#
#eBPF application that parses HTTP packets 
#and extracts (and prints on screen) the URL contained in the GET/POST request.
#
#eBPF program http_filter is used as SOCKET_FILTER attached to eth0 interface.
#only packet of type ip and tcp containing HTTP GET/POST are returned to userspace, others dropped
#
#python script uses bcc BPF Compiler Collection by iovisor (https://github.com/iovisor/bcc)
#and prints on stdout the first line of the HTTP GET/POST request containing the url

from __future__ import print_function
from bcc import BPF

import os
import socket
import struct
import sys

if len(sys.argv) != 2:
  print("usage: %s <iface>" % sys.argv[0])
  sys.exit(1)

iface = sys.argv[1]

# initialize BPF - load source code from http-parse-simple.c
bpf = BPF(src_file = "http-parse-simple.c",debug = 0)

#load eBPF program http_filter of type SOCKET_FILTER into the kernel eBPF vm
#more info about eBPF program types
#http://man7.org/linux/man-pages/man2/bpf.2.html
function_http_filter = bpf.load_func("http_filter", BPF.SOCKET_FILTER)

#create raw socket, bind it to the interface
#attach bpf program to socket created
# TODO: attach this to all interfaces
BPF.attach_raw_socket(function_http_filter, iface)

#get file descriptor of the socket previously created inside BPF.attach_raw_socket
socket_fd = function_http_filter.sock

#create python socket object, from the file descriptor
sock = socket.fromfd(socket_fd,socket.PF_PACKET,socket.SOCK_RAW,socket.IPPROTO_IP)
#set it as blocking socket
sock.setblocking(True)

while 1:
  #retrieve raw packet from socket
  packet_str = os.read(socket_fd,2048)

  #DEBUG - print raw packet in hex format
  # print ("%s" % packet_str.encode("hex"))

  #convert packet into bytearray
  packet_bytearray = bytearray(packet_str)
  
  #ethernet header length
  ETH_HLEN = 14 

  #IP HEADER
  #https://tools.ietf.org/html/rfc791
  # 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
  # +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
  # |Version|  IHL  |Type of Service|          Total Length         |
  # +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
  #
  #IHL : Internet Header Length is the length of the internet header 
  #value to multiply * 4 byte
  #e.g. IHL = 5 ; IP Header Length = 5 * 4 byte = 20 byte
  #
  #Total length: This 16-bit field defines the entire packet size, 
  #including header and data, in bytes.

  #calculate packet total length
  total_length = packet_bytearray[ETH_HLEN + 2]               #load MSB
  total_length = total_length << 8                            #shift MSB
  total_length = total_length + packet_bytearray[ETH_HLEN+3]  #add LSB
  
  #calculate ip header length
  ip_header_length = packet_bytearray[ETH_HLEN]               #load Byte
  ip_header_length = ip_header_length & 0x0F                  #mask bits 0..3
  ip_header_length = ip_header_length << 2                    #shift to obtain length

  src_ip = ".".join(map(lambda x: str(int(x)), packet_bytearray[ETH_HLEN + 12 : ETH_HLEN + 16]))
  dst_ip = ".".join(map(lambda x: str(int(x)), packet_bytearray[ETH_HLEN + 16 : ETH_HLEN + 20]))

  #TCP HEADER 
  #https://www.rfc-editor.org/rfc/rfc793.txt
  #  12              13              14              15  
  #  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 
  # +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
  # |  Data |           |U|A|P|R|S|F|                               |
  # | Offset| Reserved  |R|C|S|S|Y|I|            Window             |
  # |       |           |G|K|H|T|N|N|                               |
  # +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
  #
  #Data Offset: This indicates where the data begins.  
  #The TCP header is an integral number of 32 bits long.
  #value to multiply * 4 byte
  #e.g. DataOffset = 5 ; TCP Header Length = 5 * 4 byte = 20 byte

  #calculate tcp header length
  tcp_header_length = packet_bytearray[ETH_HLEN + ip_header_length + 12]  #load Byte
  tcp_header_length = tcp_header_length & 0xF0                            #mask bit 4..7
  tcp_header_length = tcp_header_length >> 2                              #SHR 4 ; SHL 2 -> SHR 2

  # Print the src and destination IPs
  src_port = struct.unpack_from('!H', packet_bytearray, ETH_HLEN + ip_header_length)[0]
  dst_port = struct.unpack_from('!H', packet_bytearray, ETH_HLEN + ip_header_length + 2)[0]
  print("%s %d %s %d" % (src_ip, src_port, dst_ip, dst_port))
  sys.stdout.flush()
