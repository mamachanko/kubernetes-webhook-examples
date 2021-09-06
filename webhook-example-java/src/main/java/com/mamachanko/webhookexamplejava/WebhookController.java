package com.mamachanko.webhookexamplejava;

import io.kubernetes.client.admissionreview.models.AdmissionResponse;
import io.kubernetes.client.admissionreview.models.AdmissionReview;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

@RestController
class WebhookController {

	private final Logger logger = LoggerFactory.getLogger(this.getClass());

	@PostMapping("/webhooks/admission/allow-all")
	public AdmissionReview admitAll(@RequestBody AdmissionReview admissionReview) {
		logger.info("Admitting {}", admissionReview);
		AdmissionResponse response = new AdmissionResponse();
		response.setAllowed(true);
		admissionReview.setResponse(response);
		return admissionReview;
	}
}
