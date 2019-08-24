import React from 'react';


function ErrorDisplay() {
    return <p>Unable to sign in</p>;
}

function p2(){
    fetch('http://localhost:8000/s/motion', {
        method: 'GET',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        },
        credentials: 'include'
       
    })
    .then((res) => {
        console.log(res.status)
    })

}

class LoginForm extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            username: '',
            password: ''
        };
        this.error = false;
        this.handleChange = this.handleChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
    }



    handleChange(event) {
        this.setState({
            [event.target.name]: event.target.value
        });
    }

    handleSubmit(event) {

        fetch('http://localhost:8000/login', {
            method: 'POST',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json',
            },
            credentials: 'include',
            body: JSON.stringify({
                username: this.state.username,
                password: this.state.password,
            })
        })
        .then((res) => {
            console.log(res.status)
            p2()
        })


       
        this.setState({ username: "no" })
        alert('A name was submitted: ' + this.state.value);
        event.preventDefault();
    }

    render() {
        return (
            <div class="w3-display-topmiddle">
                { this.error &&
                    <ErrorDisplay />
                }
                <form class="w3-container" onSubmit={this.handleSubmit}>
                    <div class="w3-section">
                        <label><b>Username</b></label>
                        <input class="w3-input w3-border w3-margin-bottom" name="username" type="text" value={this.state.username} onChange={this.handleChange} placeholder="Username" />

                        <label><b>Password</b></label>
                        <input class="w3-input w3-border w3-margin-bottom" name="password" type="password" value={this.state.password} onChange={this.handleChange} placeholder="Password" />
                        <button class="w3-button w3-block w3-green w3-section w3-padding" type="submit">Login</button>
                    </div>
                </form>
            </div>

        );
    }
}


export default LoginForm