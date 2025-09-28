import { useForm } from "@tanstack/react-form";

import './Styles/signup.css'
import motchi_pixel_logo from '../assets/motchi_pixel_logo.svg'


function Signup() {

    interface RegistrationFormValues {
        username: string;
        email: string;
        password: string;
        confirm_password: string;
    }

    const form = useForm({
        defaultValues: {
            username: '',
            email: '',
            password: '',
            confirm_password: '',
        } as RegistrationFormValues,
        onSubmit: ({value}) => {
            const user: RegistrationFormValues = {
                username: value.username,
                email: value.email,
                password: value.password,
                confirm_password: value.confirm_password
            }
            console.log(value)
            // alert(JSON.stringify(value, null, 2))
        },
    })

    return (
        <>
            <main className='login-cont'>
                <header>
                    <img src={`${motchi_pixel_logo}`} alt="logo for motchi" className="logo" />
                    <h1>Sign Up</h1>
                </header>
                    <form className='formContainer' action="" 
                    onSubmit={(e) => {
                        e.preventDefault();
                        form.handleSubmit();}}>
                    {/* Username */}
                    <div className='form'>
                        <form.Field 
                            name='username'
                            validators={{
                                onChange: ({ value }) => {
                                    return value.trim() === "" ? "Username is required" : undefined
                                }
                            }}
                            children={(field) => (
                                <div className='field'>
                                    <input 
                                        placeholder='Username'
                                        type='text'
                                        id={ field.name }
                                        value={ field.state.value }
                                        onBlur={ field.handleBlur }
                                        onChange={ (e) => field.handleChange(e.target.value) } 
                                    />
                                    {field.state.meta.errors.length > 0 && (
                                        <p className='warning'>{field.state.meta.errors.join(", ")}</p>
                                    )}
                                </div>
                            )}
                        />
                        {/* Email */}
                        <form.Field 
                            name='email'
                            validators={{
                                onChange: ({ value }) => {
                                    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
                                    return !emailRegex.test(value) ? "Please enter a valid email." : undefined;
                                }
                            }}
                            children={(field) => (
                                <div className='field'>
                                    <input type="email"
                                            placeholder='Email'
                                            id={ field.name }
                                            value={ field.state.value }
                                            onBlur={ field.handleBlur }
                                            onChange={ (e) => field.handleChange(e.target.value) }
                                    />
                                    {field.state.meta.errors.length > 0 && (
                                        <p className='warning'>{field.state.meta.errors.join(", ")}</p>
                                    )}
                                </div>
                            )}
                        />
                        {/* Password */}
                        <form.Field     
                            name='password'
                            validators={{
                                onChange: ({ value }) => {
                                    return value.length < 10 ? 'Password must be at least 10 characters' : undefined
                                }
                            }}
                            children={(field) => (
                                <div className='field'>
                                    <input type="password"
                                            placeholder='Password'
                                            id={field.name}
                                            name={field.name}
                                            value={field.state.value}
                                            onBlur={field.handleBlur}
                                            onChange={(e) => field.handleChange(e.target.value)}
                                    />
                                    {field.state.meta.errors.length > 0 && (
                                        <p className='warning'>{field.state.meta.errors.join(", ")}</p>
                                    )}
                                </div>
                            )}
                        />  
                        {/* Confirm Password */}
                        <form.Field     
                            name='confirm_password'
                            validators={{
                                onChangeListenTo: ['password'],
                                onChange: ({ value, fieldApi }) => {
                                    return value !== fieldApi.form.getFieldValue('password') ? 'Passwords do not match.' : undefined
                                }
                            }}
                            children={(field) => (
                                <div className='field'>
                                    <input type="password"
                                            placeholder='Confirm Password'
                                            id={field.name}
                                            name={field.name}
                                            value={field.state.value}
                                            onBlur={field.handleBlur}
                                            onChange={(e) => field.handleChange(e.target.value)}
                                    />
                                    {field.state.meta.errors.length > 0 && (
                                        <p className='warning'>{field.state.meta.errors.join(", ")}</p>
                                    )}
                                </div>
                            )}
                        /> 
                    </div>
                    <button type='submit' className='CTA'>Sign Up</button>
                </form>
            </main>
        </>
    );
}

export default Signup;