# Requirements
*This application is written in Golang v1.21. It was developed in an Ubuntu 22.04 LTS environment.*

Optional installations other than Golang are:
- [Miniconda](https://docs.conda.io/projects/miniconda/en/latest/)
    + Hash of image I used: d0643508fa49105552c94a523529f4474f91730d3e0d1f168f1700c43ae67595
- Make (sudo apt install -y make)

# Prepare project structure
1. Place the dataset CSVs in a folder, eg: `./ml-latest`
    - Example:
    `tree -d .`
    ```
    .
    ├── ml-latest
    └── recommender
        ├── algorithms
        ├── config
        ├── helpers
        ├── models
        ├── preprocess
        ├── tests
        ├── ui
        │   ├── css
        │   └── js
        └── utils
    ```
2. From now on, working directory is `recommender`
    - `cd ./recommender`

# Compilation & Execution
* If you installed miniconda:
    1. `conda create --name go --channel conda-forge go=1.21`
    2. `conda activate go`

* Compile & run each the binaries at once with:
    1. Preprocess: `go run preprocess/preprocess.go -d ../ml-latest -p preprocessed-data`
    2. Recommender: `go run recommender -d preprocessed-data -n 100 -s cosine -a tag -i 6539`
        - Note that there is also an optional parameter `-r maxRecords` which limits the dataset depending on the algorithm.
            + Sample usage: `go run recommender -d preprocessed-data -n 100 -s cosine -a item -i 1 -r 5000`
            + This uses the first `maxRecords` objects in the dataset, eg. the first 5000 movies with *all* their ratings in the above case.
    3. UI: `go run recommender -d preprocessed-data -u`
        - *Note: Preprocess needs to be executed at least once before recommender to produce the following files:*
        ``` 
            preprocessed-data
            ├── movieTitles.gob
            ├── movies.gob
            ├── tags.gob
            └── users.gob
        ```
        - The optional parameter `maxRecords` can be specified through the UI as well.

* Alternativelly if you want to seperate compilation and execution steps do one of the following:
    - If you have make installed you can run `make` which will build `recommender` and `preprocess/preprocess` binaries
    - If you don't have make installed, compile with:
        + Preprocess: `go build -o preprocess/preprocess preprocess/preprocess.go`
        + Recommender: `go build -o recommender .`
    - Then execute with:
    1. Preprocess: `./preprocess/preprocess -d ../ml-latest -p preprocessed-data`
    2. Recommender: `./recommender -d preprocessed-data -n 100 -s cosine -a tag -i 6539`
    3. UI: `./recommender -d preprocessed-data -u`

# Detailed report
You can find the detailed report regarding the implementation of my recommender app in `./Report.pdf`
