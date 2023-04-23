const Controller = {
  search: (ev) => {
    ev.preventDefault();
    const form = document.getElementById("form");
    const data = Object.fromEntries(new FormData(form));
    const response = fetch(`/search?q=${data.query}`).then((response) => {
      response.json().then((results) => {
        Controller.updateTable(results, data.query);
      });
    });
  },

  updateTable: (results, query) => {
    const table = document.getElementById("table-body");
    const tableHeader = document.getElementById("table-header");
    const rows = [];
    rows.push(`<tr>
      <th scope="col">Chapter</th>
      <th scope="col", style="text-align:left">Excerpt</th>
    </tr>`)
    for (let result of results) {
      [chapter, resString] = result
      rows.push(`<tr>
        <th scope="row", style="vertical-align:top">${chapter}</th>
        <td>${Controller.updateResult(resString, query)}</td>
      </tr>`)
    }
    table.innerHTML = rows;
  },

  updateResult: (result, query) => {
    result = Controller.formatResultString(result)
    queryLower = query.toLowerCase();
    queryUpper = query.toUpperCase();
    queryCapitalize = queryLower.charAt(0).toUpperCase() + queryLower.slice(1);
    result = Controller.highlightSearchTerm(result, queryLower)
    result = Controller.highlightSearchTerm(result, queryUpper)
    result = Controller.highlightSearchTerm(result, queryCapitalize)

    return result

  },

  checkForRomanNumeral: (text) => {
    romanNumeralRegex = /^M{0,3}(CM|CD|D?C{0,3})(XC|XL|L?X{0,3})(IX|IV|V?I{0,3}[.]?)$/
    match = text.match(romanNumeralRegex);
    if (match == null){ return false }
    return true
  },

  createSeparateLine: (text) => {
    return '<p>' + text + '<\p>';
  },



  formatResultString: (result) => {
    allCapLettersRegex = /(\b[A-Z]{2,}\b[,.]?)/
    const newText = result.split(allCapLettersRegex)

    newResult = ""
    i = 0
    while(i < newText.length){
      if (newText[i] == 'ACT' || newText[i] == 'SCENE' ){
        if (i + 2 < newText.length && Controller.checkForRomanNumeral(newText[i+1])){
          newResult += Controller.createSeparateLine(`${newText[i]} ${newText[i+1]} ${newText[i+2]}`);
          i = i + 3
          continue
        }
        if (i + 1 < newText.length) {
          newResult += Controller.createSeparateLine(`${newText[i]} ${newText[i+1]}`);
          i = i + 2
          continue
        }
      }

      if (i + 1 < newText.length && Controller.checkForRomanNumeral(newText[i+1])){
        newResult += Controller.createSeparateLine(`${newText[i]} ${newText[i+1]}`);
        i = i + 2
        continue
      }

      newResult += '<p>' + newText[i] + '<\p>';
      i++
    }

    return newResult
  },

  highlightSearchTerm: (result, query) => {
    return result.replaceAll(query, '<span style="background-color:#ffbf00;color:#fff;"><b>' + query + '</b></span>')
  }
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);
