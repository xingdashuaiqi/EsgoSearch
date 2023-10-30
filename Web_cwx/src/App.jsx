import React, { Component } from 'react';

class SearchComponent extends Component {
  constructor() {
    super();
    this.state = {
      query: '',
      results: [],
    };
  }

  // 在 SearchComponent 类中添加一个方法用于截取内容并高亮关键字
  highlightAndTruncateContent = (content, query) => {
    const keywordIndex = content.toLowerCase().indexOf(query.toLowerCase());
    const startIndex = Math.max(0, keywordIndex - 150);
    const endIndex = Math.min(content.length, keywordIndex + query.length + 150);
  
    if (keywordIndex === -1) {
      return content.substring(0, 300);
    }
  
    const truncatedContent = content.substring(startIndex, endIndex);
  
    const highlightedContent = truncatedContent.replace(
      new RegExp(query, 'gi'),
      match => `<mark>${match}</mark>`
    );
  
    return highlightedContent;
  };

  handleSubmit = async (e) => {
    e.preventDefault();
    const { query } = this.state;
    if (query.trim() !== '') {
      try {
        const response = await fetch(`http://localhost:8080/search?query=${encodeURIComponent(query)}`);
        if (!response.ok) {
          throw new Error(`Request failed with status ${response.status}`);
        }
        const data = await response.json();
        console.log('Response from server:', data);
        this.setState({ results: data });
      } catch (error) {
        console.error('Error fetching search results:', error);
      }
    }
  };

  render() {
    const { query, results } = this.state;

    const styles = {
      body: {
        fontFamily: 'Arial, Helvetica, sans-serif',
        backgroundColor: '#f5f5f5',
        margin: 0,
        padding: '20px',
        height: '100vh',
        width: '100vw',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
      },
      h1: {
        backgroundColor: '#3385ff',
        color: 'white',
        width: '100%',
        padding: '20px',
        margin: 0,
        fontSize: '24px',
      },
      form: {
        textAlign: 'center',
        marginTop: '20px',
      },
      input: {
        padding: '10px',
        fontSize: '18px',
        border: '2px solid #ccc',
        borderRadius: '4px',
        width: '60%',
        marginBottom: '10px',
      },
      button: {
        padding: '12px 24px',
        fontSize: '18px',
        backgroundColor: '#3385ff',
        color: 'white',
        border: 'none',
        borderRadius: '4px',
        cursor: 'pointer',
      },
      h2: {
        marginTop: '20px',
        paddingLeft: '20px',
        fontSize: '20px',
      },
      ul: {
        listStyle: 'none',
        padding: 0,
      },
      li: {
        backgroundColor: 'white',
        border: '1px solid #ccc',
        borderRadius: '4px',
        margin: '10px',
        padding: '20px',
        boxShadow: '0 4px 8px rgba(0, 0, 0, 0.2)',
      },
      a: {
        textDecoration: 'none',
        color: '#3385ff',
      },
      noResults: {
        textAlign: 'center',
        marginTop: '20px',
        color: '#666',
        fontSize: '18px',
      },
    };

    return (
      <div style={styles.body}>
        <h1 style={styles.h1}>专属于你的超级搜索</h1>
        <form onSubmit={this.handleSubmit} id="search-form" style={styles.form}>
          <label htmlFor="query">请输入搜索关键字：</label>
          <input
            type="text"
            id="query"
            name="query"
            value={query}
            onChange={(e) => this.setState({ query: e.target.value })}
            required
            style={styles.input}
          />
          <button
            type="submit"
            style={styles.button}
          >
            搜索
          </button>
        </form>
        <h2 style={styles.h2}>搜索结果：</h2>
        <ul id="results-list" style={styles.ul}>
          {results.length > 0 ? (
            results.map((result, index) => (
              <li key={index} style={styles.li}>
                <a href={result.url} style={styles.a}>
                  <strong>{result.title}</strong> - {result.date}
                  <br />
                  <div
                    dangerouslySetInnerHTML={{
                      __html: this.highlightAndTruncateContent(result.content, query),
                    }}
                  />
                </a>
              </li>
            ))
          ) : (
            <li className="no-results" style={styles.noResults}>未找到相关结果</li>
          )}
        </ul>
      </div>
    );
  }
}

export default SearchComponent;
