
import styles from './Styles/login.module.css'

import { NavLink } from "react-router-dom";

import { useState } from "react";
import { useForm } from '@tanstack/react-form'
import motchi_pixel_logo from "../assets/motchi_pixel_logo.svg"
import { MdOutlineVisibility } from "react-icons/md";
import { MdOutlineVisibilityOff } from "react-icons/md";


interface RegistrationFormValues {
    username: string;
    email: string;
    password: string;
    confirm_password: string;
}

export function Login() {

    const [showPassword, setShowPassword] = useState(false);

    const togglePasswordVisibility = () => {
    setShowPassword(!showPassword);
    };

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
            alert(JSON.stringify(value, null, 2))
        },
    })

  return (
    <>
        <main className={`${styles.loginCont}`}>
            <header>
                <img src={`${motchi_pixel_logo}`} alt="logo for motchi" className={`${styles.logo}`} />
                <h1>Sign In</h1>
            </header>
            <form className={`${styles.formContainer}`} action="" 
                        onSubmit={(e) => {
                            e.preventDefault();
                            form.handleSubmit();}}
                    >
                    <div className={`${styles.form}`}>
                        {/* Username */}
                        <form.Field
                            name='username'
                            validators={{
                                onChange: ({ value }) => {
                                    return value.trim() === "" ? "Enter username" : undefined
                                }
                            }}
                            children={(field) => (
                                <div className={`${styles.field}`}>
                                    <input 
                                        placeholder='Username'
                                        type='text'
                                        id={ field.name }
                                        value={ field.state.value }
                                        onBlur={ field.handleBlur }
                                        onChange={ (e) => field.handleChange(e.target.value) } 
                                    />
                                    {field.state.meta.errors.length > 0 && (
                                        <p className={`${styles.warning}`}>{field.state.meta.errors.join(", ")}</p>
                                    )}
                                </div>
                                )}
                        />
                        {/* Password */}
                        <form.Field     
                            name='password'
                            validators={{
                                onChange: ({ value }) => {
                                    return value.trim() === "" ? "Enter password" : undefined
                                }
                            }}
                            children={(field) => (
                                <>
                                    <div className={`${styles.field} ${styles.passwordCont}`}>
                                        <input type={showPassword ? 'text' : 'password'}
                                                placeholder='Password'
                                                id={field.name}
                                                name={field.name}
                                                value={field.state.value}
                                                onBlur={field.handleBlur}
                                                onChange={(e) => field.handleChange(e.target.value)}
                                        />
                                        <button className={`${styles.iconButton}`}
                                                onClick={togglePasswordVisibility}>
                                            {showPassword ? <MdOutlineVisibilityOff /> : <MdOutlineVisibility />}
                                        </button>
                                    </div>
                                    <div>
                                        {field.state.meta.errors.length > 0 && (
                                            <p className={`${styles.warning}`}>{field.state.meta.errors.join(", ")}</p>
                                        )}
                                    </div>
                                </>
                            )}
                        />  
                    </div>
                    <button type='submit' className={`${styles.CTA}`}>Log in</button>

                    <div className={`${styles.register}`}>
                        <p>Don't have an account?</p>
                        <NavLink to='/sign-up' className={`${styles.registerLink}`}>
                            <p>Register</p>
                        </NavLink>
                    </div>
                </form>
        </main>
    </>
  );
}