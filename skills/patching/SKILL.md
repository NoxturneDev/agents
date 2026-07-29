---
name: patching
description: Automates fleet deployments, backend/frontend updates, migration & seeder detection, runtime config handling, and database patching across Ziad physical servers using Ansible and patch-local.sh.
---

# Ziad Fleet Patching & Deployment Skill

This skill provides step-by-step instructions and command references for building, distributing, and deploying updates across the physical server fleet (`server1` through `server10`) using Ansible and `patch-local.sh`.

---

## 1. Inventory & Server Mapping

Fleet configurations are maintained in [inventory.ini](file:///home/noxturne/projects/ziad/automation/ansible/inventory.ini).

| Host Group | Tailscale IP | Clients Hosted |
| :--- | :--- | :--- |
| **server1_apps** | `100.126.134.42` | `demo2`, `kis`, `madinahalhijrah`, `sdpasshiddiqiyah2`, `shohibulbarokah` |
| **server2_apps** | `100.85.196.93` | `yapira`, `setiagama`, `elazam`, `yabni`, `alkautsarsmart`, `asysyifa`, `daarelhikam`, `banitamim` |
| **server3_apps** | `100.118.103.48` | `darulhuda`, `darurrosyid`, `binka`, `albadar`, `ppmbq`, `ibadurrahman`, `albayyinah` |
| **server4_apps** | `100.109.47.68` | `almuhajirin`, `mthq`, `fajrulkarim` |
| **server5_apps** | `100.109.132.7` | `daarelkhaer`, `ponpesmodernayyusufiah`, `tiaraaksara`, `alistiqomah` |
| **server6_apps** | `100.115.194.83` | `albahjahbogor`, `albahjahcianjur`, `amalulummah`, `daarelqurro`, `darulhikmah` |
| **server7_apps** | `100.71.63.87` | `kebuntumbuh`, `alkautsarbogor`, `permataihsan`, `arrohmahbogor`, `asshofa` |
| **server8_apps** | `100.75.28.89` | `annihayah`, `ppms`, `norma`, `stitdaarulfatah`, `albahjahannahltangerang`, `ahekppulo`, `mabos` |
| **server9_apps** | `100.97.200.48` | `demo`, `demo3` |
| **server10_apps**| `100.95.50.24` | `smkfarmasitangerang1`, `bimbelmontessori`, `jagatarsy`, `daarelilmi`, `albahjahpekanbaru` |

---

## 2. Pre-Deployment Reconnaissance (Mandatory)

Always perform these 3 checks **before** running a deployment:

### Step 2.1: Check Live Server Tags
```bash
ansible <host|group> -i inventory.ini -m shell -a "git -C {{ api_dir }} describe --tags --abbrev=0" --become
```

### Step 2.2: Check for Uncommitted Working Copy Changes
```bash
ansible <host|group> -i inventory.ini -m shell -a "git -C {{ api_dir }} status --porcelain" --become
```
*Note: If modified files (`M`) exist on a target client, resolve or stash them before running `git checkout`.*

### Step 2.3: Detect Pending Migrations & Seeders
Fetch latest tags in backend repository and run diff:
```bash
git -C /home/noxturne/projects/ziad/backend/ziad-laravel-template fetch --tags
git -C /home/noxturne/projects/ziad/backend/ziad-laravel-template diff --name-only <FROM_TAG> <TARGET_TAG> | grep -E "database/migrations|database/seeders"
```

---

## 3. Deployment Workflow & Command Patterns

All deployment commands are run from directory `/home/noxturne/projects/ziad/automation/ansible`.

### Pattern A: Standard Deployment (With Local Bun Build)
```bash
./patch-local.sh -v <TARGET_TAG> -l <HOST_OR_GROUP> -m -s <SeederClass> --no-docker -y
```

### Pattern B: Fast Reuse Deployment (`--skip-build`)
*Use when `dist/` is already compiled from a previous build on the same release tag. Reuses `dist/` instantly.*
```bash
./patch-local.sh -v <TARGET_TAG> -l <HOST_OR_GROUP> --skip-build -m -s <SeederClass> --no-docker -y
```

### Pattern C: Backend-Only Deployment (`--skip-fe`)
*Use when only PHP/Laravel code changed and frontend does not need updating.*
```bash
./patch-local.sh -v <TARGET_TAG> -l <HOST_OR_GROUP> --skip-fe -m -y
```

---

## 4. Database Patching via Artisan Tinker (Non-Raw MySQL)

To run database updates across clients safely using Laravel Eloquent / Query Builder instead of raw SQL:

### Step 4.1: Record Count Check (Dry Run)
```bash
ansible <host|group> -i inventory.ini -m shell -a 'php {{ api_dir }}/artisan tinker --execute="echo \Illuminate\Support\Facades\DB::table(\"<table_name>\")->where(\"<column>\", \"like\", \"<pattern>%\")->count();"' --become
```

### Step 4.2: Execute Update via Query Builder
```bash
ansible <host|group> -i inventory.ini -m shell -a 'php {{ api_dir }}/artisan tinker --execute="\Illuminate\Support\Facades\DB::table(\"<table_name>\")->where(\"<column>\", \"like\", \"<pattern>%\")->update([\"<column_to_set>\" => <value>]);"' --become
```

---

## 5. Post-Deployment Verification

Always confirm the deployed tag on the target servers after execution:
```bash
ansible <host|group> -i inventory.ini -m shell -a "git -C {{ api_dir }} describe --tags --abbrev=0" --become
```

---

## 6. Important Safety Rules

1. **Ask for Confirmation:** Always present the pre-deployment analysis (current tag, new migrations, seeders, and planned commands) to the user before running deployments.
2. **Preserve Runtime Config:** Never overwrite `/var/www/<domain>/config.js` on remote servers. The deployment playbook automatically excludes `config.js` via `rsync`.
3. **Use `--no-docker`:** Always use `--no-docker` when running `./patch-local.sh` to compile using the host's local `bun` binary.
