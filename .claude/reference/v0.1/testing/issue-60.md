# Issue #60 Test Log — Add --authored flag to records list (ORCID-based search)

**Result: ALL PASS (15/15)**

Date: 2026-02-20
Branch: `feat/authored-flag`
PR: #62

---

## Test 1: `./zenodo.exe config list`

**Status: PASS**

```
profile:     production
base_url:    https://zenodo.org/api
token:       ****6DVp
orcid:       0000-0002-4699-4755
```

## Test 2: `./zenodo.exe config delete orcid`

**Status: PASS**

```
Deleted orcid
```

## Test 3: `./zenodo.exe records list` (before setting ORCID)

**Status: PASS** — Shows self-uploaded records with hint to set ORCID.

```
Records uploaded by your account (to see all records you authored or contributed to: zenodo config set orcid <your-orcid>)
Showing 6 records
              TITLE              |             COMMUNITY             |              DOI               |  CREATED
---------------------------------+-----------------------------------+--------------------------------+-------------
  SALURBAL Project Portal - s... | salurbal                          | https://doi.org/10.5281/zen... | 2025-12-10
  Philadelphia Council Distri... | drexel-uhc                        | https://doi.org/10.5281/zen... | 2025-08-12
  Bringing public health to t... | salurbal                          | https://doi.org/10.5281/zen... | 2025-07-25
  CDC Social Vulnerability In... | drexel-uhc                        | https://doi.org/10.5281/zen... | 2025-02-05
  SALURBAL - FAIR primer         | drexel-uhc                        | https://doi.org/10.5281/zen... | 2024-02-19
  Drexel UHC Summer Institute... | drexel-uhc, eu, worldfair-project | https://doi.org/10.5281/zen... | 2023-11-30
```

## Test 4: `./zenodo.exe config set orcid 0000-0002-4699-4755`

**Status: PASS**

```
Set orcid = 0000-0002-4699-4755
```

## Test 5: `./zenodo.exe config get orcid`

**Status: PASS**

```
0000-0002-4699-4755
```

## Test 6: `./zenodo.exe config list` (orcid should show value)

**Status: PASS**

```
profile:     production
base_url:    https://zenodo.org/api
token:       ****6DVp
orcid:       0000-0002-4699-4755
```

## Test 7: `./zenodo.exe records list` (after setting ORCID)

**Status: PASS** — Shows authored/contributed records with stats.

```
Records where you are a creator or contributor (ORCID 0000-0002-4699-4755)
Showing 6 of 6 records
         TITLE         |             COMMUNITY             |         DOI          | VIEWS | DOWNLOADS |  CREATED
-----------------------+-----------------------------------+----------------------+-------+-----------+-------------
  SALURBAL Project ... | salurbal                          | https://doi.org/1... |    63 |        15 | 2025-12-10
  Bringing public h... | salurbal                          | https://doi.org/1... |   170 |       373 | 2025-07-25
  SALURBAL - FAIR p... | drexel-uhc                        | https://doi.org/1... |    48 |        12 | 2024-02-19
  Philadelphia Coun... | drexel-uhc                        | https://doi.org/1... |   130 |        85 | 2025-08-12
  Drexel UHC Summer... | drexel-uhc, eu, worldfair-project | https://doi.org/1... |   517 |      1143 | 2023-11-30
  CDC Social Vulner... | drexel-uhc                        | https://doi.org/1... |   296 |       293 | 2025-02-05
```

## Test 8: `./zenodo.exe records list --authored`

**Status: PASS** — Same output as default with ORCID set.

```
Records where you are a creator or contributor (ORCID 0000-0002-4699-4755)
Showing 6 of 6 records
```

## Test 9: `./zenodo.exe records list --authored --community=salurbal`

**Status: PASS** — Filters to 2 salurbal records.

```
Records where you are a creator or contributor (ORCID 0000-0002-4699-4755)
Showing 2 of 2 records
              TITLE              | COMMUNITY |              DOI               | VIEWS | DOWNLOADS |  CREATED
---------------------------------+-----------+--------------------------------+-------+-----------+-------------
  SALURBAL Project Portal - s... | salurbal  | https://doi.org/10.5281/zen... |    63 |        15 | 2025-12-10
  Bringing public health to t... | salurbal  | https://doi.org/10.5281/zen... |   170 |       373 | 2025-07-25
```

## Test 10: `./zenodo.exe records list --authored --status draft` (should error)

**Status: PASS** — Errors as expected.

```
{"code":1,"error":"--authored cannot be used with --status draft (drafts are not available via the search API)"}
EXIT: 1
```

## Test 11: `./zenodo.exe records list --community` (aggregate)

**Status: PASS** — Aggregates across 2 communities.

```
Total: 6 records across 2 communities
  COMMUNITY  |             TITLE             |              DOI              | VIEWS | DOWNLOADS |  CREATED
-------------+-------------------------------+-------------------------------+-------+-----------+-------------
  salurbal   | SALURBAL Project Portal - ... | https://doi.org/10.5281/ze... |    63 |        15 | 2025-12-10
  salurbal   | Bringing public health to ... | https://doi.org/10.5281/ze... |   170 |       373 | 2025-07-25
  drexel-uhc | Philadelphia Council Distr... | https://doi.org/10.5281/ze... |   130 |        85 | 2025-08-12
  drexel-uhc | CDC Social Vulnerability I... | https://doi.org/10.5281/ze... |   296 |       293 | 2025-02-05
  drexel-uhc | SALURBAL - FAIR primer        | https://doi.org/10.5281/ze... |    48 |        12 | 2024-02-19
  drexel-uhc | Drexel UHC Summer Institut... | https://doi.org/10.5281/ze... |   517 |      1143 | 2023-11-30
```

## Test 12: `./zenodo.exe records list --community=salurbal`

**Status: PASS** — Shows 2 salurbal records.

```
Showing 2 of 2 records in salurbal
  COMMUNITY |             TITLE              |              DOI               | VIEWS | DOWNLOADS |  CREATED
------------+--------------------------------+--------------------------------+-------+-----------+-------------
  salurbal  | SALURBAL Project Portal - s... | https://doi.org/10.5281/zen... |    63 |        15 | 2025-12-10
  salurbal  | Bringing public health to t... | https://doi.org/10.5281/zen... |   170 |       373 | 2025-07-25
```

## Test 13: `./zenodo.exe config delete orcid`

**Status: PASS**

```
Deleted orcid
```

## Test 14: `./zenodo.exe config get orcid` (should error: key not found)

**Status: PASS**

```
{"code":1,"error":"key \"orcid\" not found"}
EXIT: 1
```

## Test 15: `./zenodo.exe records list` (after deleting ORCID)

**Status: PASS** — Falls back to self-uploaded records with hint.

```
Records uploaded by your account (to see all records you authored or contributed to: zenodo config set orcid <your-orcid>)
Showing 6 records
              TITLE              |             COMMUNITY             |              DOI               |  CREATED
---------------------------------+-----------------------------------+--------------------------------+-------------
  SALURBAL Project Portal - s... | salurbal                          | https://doi.org/10.5281/zen... | 2025-12-10
  Philadelphia Council Distri... | drexel-uhc                        | https://doi.org/10.5281/zen... | 2025-08-12
  Bringing public health to t... | salurbal                          | https://doi.org/10.5281/zen... | 2025-07-25
  CDC Social Vulnerability In... | drexel-uhc                        | https://doi.org/10.5281/zen... | 2025-02-05
  SALURBAL - FAIR primer         | drexel-uhc                        | https://doi.org/10.5281/zen... | 2024-02-19
  Drexel UHC Summer Institute... | drexel-uhc, eu, worldfair-project | https://doi.org/10.5281/zen... | 2023-11-30
```
