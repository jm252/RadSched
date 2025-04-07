import subprocess
import re
import json

def ping_region(region):
    url = f"ec2.{region}.amazonaws.com"  # EC2 URL
    try:
        result = subprocess.run(["/sbin/ping", "-c", "1", url], capture_output=True, text=True)
        match = re.search(r"time=([\d.]+)\s*ms", result.stdout)
        rtt = float(match.group(1)) if match else None

        if rtt is not None:
            return rtt
        else:
            print(f"Received no RTT data from {url}")
            return None
    except Exception as e:
        print(f"Failed to ping {url}: {e}")
        return None

def ping_regions():
    regions = ["us-west-1", "us-east-1", "us-west-2", "us-east-2", "ap-east-1",
               "ap-south-1", "ap-northeast-1", "ap-northeast-2", "ap-northeast-3"]
    
    rtt_data = {}
    for region in regions:
        rtt = ping_region(region)
        if rtt is not None:
            rtt_data[region] = rtt
        else:
            rtt_data[region] = -1 
    
    return rtt_data

if __name__ == "__main__":
    data = ping_regions()
    print(json.dumps(data, indent=4))
