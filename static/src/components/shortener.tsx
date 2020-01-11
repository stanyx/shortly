import * as React from 'react';
import axios from 'axios';

interface ShortenerState {
    longUrl: string;
    shortUrl: string;
}

export class Shortener extends React.Component<{}, ShortenerState> {
    constructor(props: any) {
        super(props);
        this.state = {longUrl: "", shortUrl: ""};
    }
    submit(e: any) {
        e.preventDefault();
        axios.post('/api/v1/links', {
            url: this.state.longUrl
        }).then((response: any) => {
            console.log(response);
            this.setState({shortUrl: response.data.result.short})
        });
    }
    render() {
        return (
            <form>
                <div className="shortener-container">
                    <textarea value={this.state.longUrl} 
                        onChange={(e) => this.setState({longUrl: e.target.value})}
                        placeholder="Paste your long url here to get a nice short version">

                    </textarea>
                </div>
                <div className="shortener-toolbar">
                    <button className="button get-link-button" onClick={(e) => this.submit(e)}>Get Short URL</button>
                    {this.state.shortUrl ? (
                        <div className="short-link-label">Your short link:
                            <a target="_blank" href={this.state.shortUrl}>{this.state.shortUrl}</a>
                        </div>)
                        : null
                    }
                </div> 
            </form>
        );
    }
}