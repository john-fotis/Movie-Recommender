<!-- Header -->
<div id="top"/>

<h1 align="center"> Golang Movie Recommender </h1>

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li><a href="#about-the-project">About The Project</a></li>
    <li><a href="#requirements">Requirements</a></li>
    <li><a href="#prepare-project-structure">Prepare Project Structure</a></li>
    <li><a href="#compilation--execution">Compilation & Execution</a></li>
    <li><a href="#detailed-report">Detailed Report</a></li>
    <li><a href="#license">License</a></li>
  </ol>
</details>

<!-- Body -->

## About the project
This project marks my debut in learning [Go](https://go.dev/), driven by the assignment of the [Big Data](https://cgi.di.uoa.gr/~antoulas/index.html#teaching) course I attended in the context of my [Master's degree](https://www.di.uoa.gr/eng).
The application can either run in CLI or start a simple Web UI as described in [Compilation & Execution](#compilation--execution) section. The design of the current application models (eg. user, movie, rating etc.) was influenced by the input data sourced from the [grouplens latest full dataset](https://grouplens.org/datasets/movielens/latest/) as of December 2023. From this point onward, this dataset will be referenced as `ml-latest`.

## Requirements
*This application is written in Golang v1.21. It was developed in an Ubuntu 22.04 LTS environment.*

Optional installations other than Golang are:
- [Miniconda](https://docs.conda.io/projects/miniconda/en/latest/)
    + Hash of image I used: `d0643508fa49105552c94a523529f4474f91730d3e0d1f168f1700c43ae67595`
- Make (`sudo apt install -y make`)

## Prepare project structure
1. Place the dataset CSVs in a folder, eg: `./ml-latest`
    - Example:
    `tree -d .`
    ```
    .
    ├── algorithms
    ├── config
    ├── helpers
    ├── ml-latest
    ├── models
    ├── preprocess
    ├── recommenders
    ├── tests
    ├── ui
    │   ├── css
    │   └── js
    └── utils
    ```

## Compilation & Execution
* If you installed miniconda:
    1. `conda create --name go --channel conda-forge go=1.21`
    2. `conda activate go`

* Compile & run each the binaries at once with:
    1. Preprocess: `go run preprocess/preprocess.go -d ./ml-latest`
    2. Recommender: `go run recommender -n 100 -s cosine -a tag -i 6539`
        - Note that there is also an optional parameter `-r maxRecords` which limits the dataset depending on the algorithm.
            + Sample usage: `go run recommender -n 100 -s cosine -a item -i 1 -r 5000`
            + This uses the first `maxRecords` objects in the dataset, eg. the first 5000 movies with *all* their ratings in the above case.
    3. UI: `go run recommender -u`
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
    1. Preprocess: `./preprocess/preprocess -d ./ml-latest`
    2. Recommender: `./recommender -n 100 -s cosine -a tag -i 6539`
    3. UI: `./recommender -u`

## Detailed report
You can find the detailed report regarding the implementation of my recommender app in [Report.pdf](https://github.com/john-fotis/Movie-Recommender/blob/main/Report.pdf)

<!-- Footer -->

## License
This project is licensed under the [MIT License](https://github.com/john-fotis/Movie-Recommender/blob/main/LICENSE.md)
