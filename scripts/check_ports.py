import socket

TARGET_IP = "192.168.0.167"
PORTS = [3001, 3111, 5005]

def check_port(ip, port):
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.settimeout(2)
    try:
        result = s.connect_ex((ip, port))
        if result == 0:
            print(f"Port {port}: OPEN")
        else:
            print(f"Port {port}: CLOSED (Code: {result})")
    except Exception as e:
        print(f"Port {port}: ERROR ({e})")
    finally:
        s.close()

if __name__ == "__main__":
    print(f"Checking ports on {TARGET_IP}...")
    for p in PORTS:
        check_port(TARGET_IP, p)
