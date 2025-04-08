# RadSched

**RadSched** is a latency optimizing scheduler for the Radical stateful serverless edge computing system. RadSched selects the optimal edge node to invoke a function at on the basis of latency and data consistency. 

## How to Use the System

### 1a. Install Dependencies
- Go 1.20+ 
- Python 3 
- AWS CLI (with Lambda permissions)

### 1b. Run Dependencies
- Radical 
- Consistency Storage Server ([see example](https://github.com/jm252/consistency_storage_server))

### 2. Build 
```bash
go build
```

### 2. Register a Function
```bash
radsched prepare <function_name> <execution_time_ms> <primary_datacenter>
```

### 3. Bootstrap Most Up-to-Date Data (Optional)
```bash
radsched bootstrap
```

### 4. Run a Function

```bash
radsched run <function_name> --with-weight
```
---
