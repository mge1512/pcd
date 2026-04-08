calc-interest: calc-interest.cob
	cobc -x -o calc-interest calc-interest.cob

test: calc-interest
	echo -e "10000.00\n0.0350\n12" | ./calc-interest

clean: calc-interest
	rm calc-interest

