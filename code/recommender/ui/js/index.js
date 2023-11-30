document.getElementById('recommendationForm').addEventListener('submit', function (event) {
    // Disable window reload when form is submitted
    event.preventDefault();

    document.getElementById('loadingIndicator').style.display = 'block';
    document.getElementById('submitButton').setAttribute('disabled', 'disabled');

    const recommendations = parseInt(document.getElementById('recommendations').value);
    const similarity = document.getElementById('similarity').value;
    const algorithm = document.getElementById('algorithm').value;
    const input = parseInt(document.getElementById('input').value);
    const maxRecords = parseInt(document.getElementById('maxRecords').value);

    if (!isNaN(recommendations) && recommendations > 0 && !isNaN(input) && input > 0) {
        const params = {
            similarity,
            algorithm,
            recommendations,
            input,
        };
        if (!isNaN(maxRecords) && maxRecords > 0) {
            params.maxRecords = maxRecords;
        }

        const queryString = Object.keys(params)
            .filter(key => params[key] !== undefined && params[key] !== null)
            .map(key => encodeURIComponent(key) + '=' + encodeURIComponent(params[key]))
            .join('&');

        fetch('/recommend?' + queryString, {
            method: 'GET'
        })
            .then(response => {
                if (response.status !== 200) {
                    throw new Error('Network response error.');
                }
                return response.json();
            })
            .then(responseData => {
                console.log(responseData);
                document.getElementById('loadingIndicator').style.display = 'none';
                document.getElementById('submitButton').removeAttribute('disabled');
                // Clear old content if it exists
                document.getElementById('recommendationResults').innerHTML = '';
                if (responseData.message !== '') {
                    // If message is not empty, something went wrong with the query
                    document.getElementById('recommendationResults').innerText = responseData.message;
                } else {
                    // Create a table of results inside the div with id="recommendationResults"
                    const table = document.createElement('table');
                    table.classList.add('table');
                    // Create and append header columns based on the data type
                    const thead = document.createElement('thead');
                    const trHead = document.createElement('tr');
                    const thCol1 = document.createElement('th');
                    thCol1.scope = 'col';
                    thCol1.innerText = 'Movie ID';
                    const thCol2 = document.createElement('th');
                    thCol2.scope = 'col';
                    thCol2.innerText = 'Movie Title';
                    const thCol3 = document.createElement('th');
                    thCol3.scope = 'col';
                    thCol3.innerText = responseData.dataType === 'ratings' ? 'Forecasted rating' : 'Similarity';
                    trHead.appendChild(thCol1);
                    trHead.appendChild(thCol2);
                    trHead.appendChild(thCol3);
                    thead.appendChild(trHead);
                    table.appendChild(thead);
                    // Fill the table body dynamically with the responseData
                    const tbody = document.createElement('tbody');
                    responseData.data.forEach(item => {
                        // Create and append one row for each responseData object
                        const tr = document.createElement('tr');
                        const tdCol1 = document.createElement('td');
                        tdCol1.innerText = item.movieID;
                        const tdCol2 = document.createElement('td');
                        tdCol2.innerText = item.movieTitle;
                        const tdCol3 = document.createElement('td');
                        tdCol3.innerText = item.result;
                        tr.appendChild(tdCol1);
                        tr.appendChild(tdCol2);
                        tr.appendChild(tdCol3);
                        tbody.appendChild(tr);
                    });
                    table.appendChild(tbody);
                    document.getElementById('recommendationResults').appendChild(table);
                }
            })
            .catch(error => {
                console.error('Error:', error);
                document.getElementById('loadingIndicator').style.display = 'none';
                document.getElementById('submitButton').removeAttribute('disabled');
                document.getElementById('recommendationResults').innerText = 'An error occurred while fetching data.';
            });
    } else {
        alert('Recommendations and Input must be positive integers.');
        document.getElementById('loadingIndicator').style.display = 'none';
        document.getElementById('submitButton').removeAttribute('disabled');
    }
    return false;
});
