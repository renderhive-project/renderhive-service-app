import { Form, Formik, FormikConfig, FormikHelpers, FormikValues } from 'formik';
import React, { useState } from 'react'
import MultiStepFormNavigation from './MultiStepFormNavigation';
import { Step, StepLabel, Stepper } from '@mui/material';

interface Props extends FormikConfig<FormikValues> {
    children: React.ReactNode;
    showStepper: boolean;
}

// Multistep form component
const MultiStepForm = ({ children, initialValues, onSubmit, showStepper }: Props) => {
    const [stepNumber, setStepNumber] = useState(0);
    const steps = React.Children.toArray(children) as React.ReactElement[];  // get array of childrens in the form

    // new state for the form values
    const [formValues, setformValues] = useState(initialValues)

    // define some state variables
    const step = steps[stepNumber];
    const totalSteps = steps.length;
    const isLastStep = (stepNumber == totalSteps - 1);

    // methods to change step state
    const previous = (values: FormikValues) => {
        setformValues(values);
        setStepNumber(stepNumber - 1);
    }
    const next = (values: FormikValues) => {
        setformValues(values);
        setStepNumber(stepNumber + 1);
    }

    // form submission handler
    const handleSubmit = async (values: FormikValues, actions: FormikHelpers<FormikValues>) => {
        // if this form step has its own onSubmit
        // then wait until the form was submitted
        if (step.props.onSubmit) {
            await step.props.onSubmit(values);
        }

        // if this is the last form step
        if (isLastStep) {
            // make the final form submission
            return onSubmit(values, actions);
        } else {
            // otherwise, we just changed the step and we need to reset the touched state to empty object 
            actions.setTouched({});
            next(values);
        }
    };

    return (
        <div>
            <Formik
                initialValues={formValues}
                onSubmit={handleSubmit}
                validationSchema={step.props.validationSchema}
            >
                {(formik) => (
                    <Form>

                        {/* render a step visualization */}
                        { showStepper && <Stepper alternativeLabel activeStep={stepNumber} style={{marginTop: 10, marginBottom: 25}}>
                            {steps.map(currentStep => {
                                const label = currentStep.props.stepName;

                                return (
                                    <Step key={label}>
                                        <StepLabel>{label}</StepLabel>
                                    </Step>
                                )
                            })}
                        </Stepper>}

                        {/* render the actual form of this step */}
                        {step}
                        
                        {/* render the multi-step form's navigation buttons */}
                        <MultiStepFormNavigation
                            isLastStep={isLastStep}
                            hasPrevious={stepNumber > 0}
                            onBackClick={() => previous(formik.values)}
                        />    
                    </Form>
                )}
            </Formik>
        </div>
    )
}

export default MultiStepForm;
export const FormStep = ({ children }: any) => children;