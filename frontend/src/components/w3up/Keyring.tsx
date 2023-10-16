import { useEffect, useState } from 'react'
import { useKeyring } from '@w3ui/react-keyring'

export default function W3Keyring () {
  const [{ account }, { loadAgent, unloadAgent, cancelAuthorize }] = useKeyring()
  const [email, setEmail] = useState('')
  const [submitted, setSubmitted] = useState(false)

  useEffect(() => { loadAgent() }, []) // load the agent - oncevent.

  if (account) {
    return (
      <div>
        <h1>Welcome!</h1>
        <p>You are logged in as {account}!</p>
        <form onSubmit={(event: any) => { event.preventDefault(); unloadAgent() }}>
          <button type='submit'>Sign Out</button>
        </form>
      </div>
    )
  }

  if (submitted) {
    return (
      <div>
        <h1>Verify your email address!</h1>
        <p>Click the link in the email we sent to {email} to sign in.</p>
        <form onSubmit={(event: any) => { event.preventDefault(); cancelAuthorize() }}>
          <button type='submit'>Cancel</button>
        </form>
      </div>
    )
  }

  const handleAuthorizeSubmit = async (event: any) => {
    event.preventDefault()
    setSubmitted(true)
    // try {
    //   await authorize(email)
    // } catch (err) {
    //   throw new Error('failed to authorize', { cause: err })
    // } finally {
    //   setSubmitted(false)
    // }
  }

  return (
    <form onSubmit={handleAuthorizeSubmit}>
      <div>
        <label htmlFor='email'>Email address:</label>
        <input id='email' type='email' value={email} onChange={(event: any) => setEmail(event.target.value)} required />
      </div>
      <button type='submit' disabled={submitted}>Authorize</button>
    </form>
  )
}