package bosh_test

import (
	"encoding/json"
	"errors"

	"github.com/onrack/onrack-cpi/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("response handling", func() {
	Describe("exiting with an error", func() {
		It("wraps errors in a CpiResponse", func() {
			testErrMsg := "a test error"
			testErr := errors.New(testErrMsg)
			errResp := bosh.BuildErrorResponse(testErr, false, "")
			errRespBytes := []byte(errResp)

			targetResponse := bosh.CpiResponse{}
			err := json.Unmarshal(errRespBytes, &targetResponse)
			Expect(err).ToNot(HaveOccurred())

			Expect(targetResponse.Result).To(BeNil())
			Expect(targetResponse.Log).To(BeEmpty())

			targetResponseErr := targetResponse.Error
			Expect(targetResponseErr.Type).To(Equal(bosh.DefaultErrorType))
			Expect(targetResponseErr.Message).To(Equal(testErrMsg))
			Expect(targetResponseErr.Retryable).To(BeFalse())
		})
	})

	Describe("exiting successfully", func() {
		It("wraps the response in a CpiResponse", func() {
			resultMsg := "successful result"
			resultLog := "this is a log"
			resultResp := bosh.BuildResultResponse(resultMsg, resultLog)
			resultRespBytes := []byte(resultResp)
			targetResponse := bosh.CpiResponse{}

			err := json.Unmarshal(resultRespBytes, &targetResponse)
			Expect(err).ToNot(HaveOccurred())

			Expect(targetResponse.Result).To(Equal(resultMsg))
			Expect(targetResponse.Log).To(Equal(resultLog))
			Expect(targetResponse.Error).To(BeNil())
		})
	})
})
