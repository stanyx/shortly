import * as React from 'react';

class state {
    currentPage: number;
    total: number;
    offset: number;
    limit: number;
    pages: Array<number>;
    maxPage: number;
}

class Paginator extends React.Component<any, state> {
    constructor(props: any) {
        super(props);
        let {currentPage, pages, maxShowPage} = this.calculatePages();
        this.state = {
            currentPage: currentPage, 
            total: this.props.total, 
            limit: this.props.limit, 
            offset: this.props.offset,
            pages: pages,
            maxPage: maxShowPage,
        };
    }
    calculatePages() {

        let pages = [];
        let maxTotalPage = Math.ceil(this.props.total / this.props.limit);
        let currentPage = (this.props.offset / this.props.limit) + 1;
        let maxShowedPages = 9;

        let minShowPage = currentPage - (maxShowedPages - 1) / 2;
        if (minShowPage <= 0) {
            minShowPage = 1;
        }

        let maxShowPage = minShowPage + maxShowedPages;

        if (maxShowPage > (maxTotalPage - 1)) {
            maxShowPage = maxTotalPage;
        }

        minShowPage = maxShowPage - maxShowedPages;
        if (minShowPage <= 0) {
            minShowPage = 1;
        }

        for (let i = minShowPage; i <= maxShowPage; i++) {
            pages.push(i);
        }

        return {maxShowPage, currentPage, pages};
    }
    componentDidUpdate(prevProps: any) {
        if (this.props.total !== prevProps.total || 
            this.props.limit !== prevProps.limit || 
            this.props.offset !== prevProps.offset
        ) {
            let {currentPage, maxShowPage, pages} = this.calculatePages();
            this.setState({
                currentPage: currentPage, 
                total: this.props.total, 
                limit: this.props.limit, 
                offset: this.props.offset,
                pages: pages,
                maxPage: maxShowPage,
            })
        }
    }
    render() {
        return (
            <div className="row">
                <div className="col-sm-12 col-md-5">
                    <div className="dataTables_info" role="status" aria-live="polite">
                        Showing {this.state.offset + 1} to {this.state.offset + this.state.limit} of {this.state.total} entries
                    </div>
                </div>
                <div className="col-sm-12 col-md-7">
                    <div className="dataTables_paginate paging_simple_numbers">
                        <ul className="pagination">
                            <li className={"paginate_button page-item previous " + (this.state.currentPage > 1 ? '': 'disabled')}>
                                <a onClick={() => this.props.loadPage(this.state.currentPage - 1)} className="page-link">Previous</a>
                            </li>
                            {this.state.pages.map((i) => {
                                return (
                                    <li className={"paginate_button page-item " + (this.state.currentPage == i ? 'active': '')}>
                                        <a onClick={() => this.props.loadPage(i)} className="page-link">{i}</a>
                                    </li>
                                )
                            })}
                            <li className={"paginate_button page-item next " + (this.state.currentPage < this.state.maxPage ? '': 'disabled')}>
                                <a onClick={() => this.props.loadPage(this.state.currentPage + 1)} className="page-link">Next</a>
                            </li>
                        </ul>
                    </div>
                </div>
            </div>
        )
    }
}

export default Paginator;