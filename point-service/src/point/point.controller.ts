import {
  Body,
  Controller,
  Get,
  HttpException,
  HttpStatus,
  Logger,
  Post,
  Query,
} from '@nestjs/common';
import { PointService } from './point.service';
import { CreatePointDto } from './point.dto';
import { logs, SeverityNumber } from '@opentelemetry/api-logs';

const otelLogger = logs.getLogger('point-service');

@Controller('point')
export class PointController {
  private readonly logger = new Logger(PointController.name);

  constructor(private readonly pointService: PointService) {}

  @Get('calculate')
  calculatePoint(@Query('amount') amount: string) {
    this.logger.log(`GET /point/calculate request received: amount=${amount}`);
    otelLogger.emit({
      severityNumber: SeverityNumber.INFO,
      severityText: 'INFO',
      body: 'Calculate point request received',
      attributes: {
        log_type: 'business',
        event: 'calculate_point_request',
        entity_type: 'point',
        amount: amount,
      },
    });
    const parsedAmount = Number(amount);
    if (amount === undefined || amount === '' || isNaN(parsedAmount)) {
      this.logger.error(`GET /point/calculate invalid amount: ${amount}`);
      otelLogger.emit({
        severityNumber: SeverityNumber.ERROR,
        severityText: 'ERROR',
        body: 'Calculate point invalid amount',
        attributes: {
          log_type: 'business',
          event: 'calculate_point_invalid_amount',
          entity_type: 'point',
          amount: amount,
        },
      });
      throw new HttpException(
        'amount must be a number',
        HttpStatus.BAD_REQUEST,
      );
    }
    return { point: this.pointService.calculatePoint(parsedAmount) };
  }

  @Get()
  async getPoint() {
    this.logger.log('GET /point request received');
    otelLogger.emit({
      severityNumber: SeverityNumber.INFO,
      severityText: 'INFO',
      body: 'Get points request received',
      attributes: {
        log_type: 'business',
        event: 'get_points_request',
        entity_type: 'point',
      },
    });
    try {
      return await this.pointService.getPoint();
    } catch (error) {
      this.logger.error('PointService.getPoint internal error', error.stack);
      otelLogger.emit({
        severityNumber: SeverityNumber.ERROR,
        severityText: 'ERROR',
        body: 'PointService.getPoint internal error',
        attributes: { 'error.message': error.message },
      });
      throw new HttpException(error.message, HttpStatus.INTERNAL_SERVER_ERROR);
    }
  }

  @Post()
  async createPoint(@Body() body: CreatePointDto) {
    this.logger.log(
      `POST /point request received: userId=${body.userId}, orgId=${body.orgId}, amount=${body.amount}`,
    );
    otelLogger.emit({
      severityNumber: SeverityNumber.INFO,
      severityText: 'INFO',
      body: 'Deduct points request received',
      attributes: {
        log_type: 'business',
        event: 'deduct_points_request',
        entity_type: 'point',
        actor_id: body.userId,
        org_id: body.orgId,
        amount: body.amount,
      },
    });
    try {
      return await this.pointService.deductPoint(body);
    } catch (error) {
      this.logger.error('PointService.deductPoint internal error', error.stack);
      otelLogger.emit({
        severityNumber: SeverityNumber.ERROR,
        severityText: 'ERROR',
        body: 'PointService.deductPoint internal error',
        attributes: { 'error.message': error.message },
      });
      throw new HttpException(error.message, HttpStatus.INTERNAL_SERVER_ERROR);
    }
  }
}
