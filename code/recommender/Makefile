# Recommender
RECOMMENDER_DIR = .
RECOMMENDER_SRC = $(wildcard $(RECOMMENDER_DIR)/*.go)
RECOMMENDER_BIN = $(RECOMMENDER_DIR)/recommender
# Preprocess
PREPROCESS_DIR = ./preprocess
PREPROCESS_SRC = $(wildcard $(PREPROCESS_DIR)/*.go)
PREPROCESS_BIN = $(PREPROCESS_DIR)/preprocess
# Tests
TEST_DIR = ./tests

default: clean $(RECOMMENDER_BIN) $(PREPROCESS_BIN)

$(RECOMMENDER_BIN): $(RECOMMENDER_SRC)
	go build -o $@ $(RECOMMENDER_SRC)

$(PREPROCESS_BIN): $(PREPROCESS_SRC)
	go build -o $@ $(PREPROCESS_SRC)

clean:
	rm -f $(RECOMMENDER_BIN) $(PREPROCESS_BIN)

clean-preprocessed-data:
	rm -rf $(PREPROCESS_DIR)/data

test:
	go test -count=1 $(TEST_DIR)