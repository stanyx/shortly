import axios, {AxiosResponse} from 'axios';

let httpGet = (url: string): Promise<AxiosResponse> => {
    return axios.get(url, {
        headers: {
            'x-access-token': localStorage.getItem('token')
        }
    });
}

let httpPost = (url: string, body: any): Promise<AxiosResponse> => {
    return axios.post(url, body, {
        headers: {
            'x-access-token': localStorage.getItem('token')
        }
    });
}

let httpPut = (url: string, body: any): Promise<AxiosResponse> => {
    return axios.put(url, body, {
        headers: {
            'x-access-token': localStorage.getItem('token')
        }
    });
}

let httpDelete = (url: string): Promise<AxiosResponse> => {
    return axios.delete(url, {
        headers: {
            'x-access-token': localStorage.getItem('token')
        }
    });
}

export {httpGet, httpPost, httpPut, httpDelete};